package agent

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/github"
	"github.com/hejijunhao/heimdall/backend/internal/notifications"
)

// monitorLLMRate caps Claude invocations from the monitoring loop to 30 per
// minute (burst 5). This bounds API cost when the classifier falls back to
// PassthroughClassifier and every log is escalated.
var monitorLLMRate = rate.Every(2 * time.Second)

// defaultProviderName is the provider used when no app-level override is set.
const defaultProviderName = "anthropic"

type Agent struct {
	queries      *db.Queries
	providers    map[string]Provider
	config       *config.Config
	classifier   Classifier
	notifier     *notifications.Dispatcher
	githubClient *github.Client
	pipeline     *PipelineWriter
	pipelineBus  *PipelineBus
	limiter      *rate.Limiter
	mu           sync.Mutex // protects cancel and wg
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier, notifier *notifications.Dispatcher, gh *github.Client, pipeline *PipelineWriter, bus *PipelineBus) *Agent {
	providers := map[string]Provider{
		// Anthropic is always available — ANTHROPIC_API_KEY is required at config load.
		defaultProviderName: NewAnthropicProvider(cfg.AnthropicKey),
	}
	// OpenRouter is opt-in: only registered when the operator sets a key.
	// Without the key, the dropdown (Phase 4/5) won't show OpenRouter models
	// and any legacy app config requesting "openrouter" silently falls back
	// to anthropic via providerFor's default branch.
	if cfg.OpenRouterKey != "" {
		providers["openrouter"] = NewOpenRouterProvider(cfg.OpenRouterKey)
		slog.Info("openrouter provider enabled")
	}
	return &Agent{
		queries:      queries,
		providers:    providers,
		config:       cfg,
		classifier:   classifier,
		notifier:     notifier,
		githubClient: gh,
		pipeline:     pipeline,
		pipelineBus:  bus,
		limiter:      rate.NewLimiter(monitorLLMRate, 5),
	}
}

// Pipeline returns the writer used for log-pipeline events. Exposed so the
// webhook ingestion handler (which lives in the handlers package) can emit
// StageIngestion events alongside monitor.go's StageClassified / Gate /
// Assessment emits. Nil-safe: if the agent was constructed without a
// pipeline writer, callers' fire-and-forget Write* methods are no-ops.
func (a *Agent) Pipeline() *PipelineWriter {
	if a == nil {
		return nil
	}
	return a.pipeline
}

// PipelineBus returns the in-memory pub/sub hub for live pipeline events.
// Used by the Pipeline SSE handler to subscribe per-app. Nil-safe.
func (a *Agent) PipelineBus() *PipelineBus {
	if a == nil {
		return nil
	}
	return a.pipelineBus
}

// providerFor resolves a provider by name, falling back to the default
// (anthropic) if the requested name is not registered. Empty name also
// resolves to the default — this is the legacy / unconfigured path.
func (a *Agent) providerFor(name string) Provider {
	if name != "" {
		if p, ok := a.providers[name]; ok {
			return p
		}
	}
	return a.providers[defaultProviderName]
}

// Start launches the agent's background goroutines: the monitoring loop,
// the log_buffer pruner, and the investigation scheduler. Safe to call
// multiple times: stops the previous instance first. Stop() waits on a.wg,
// so all goroutines are joined before Stop returns.
func (a *Agent) Start(ctx context.Context) {
	if a.config.DisableBackgroundJobs {
		slog.Info("background jobs disabled (DISABLE_BACKGROUND_JOBS is set), skipping monitor/pruner/scheduler")
		return
	}

	a.mu.Lock()
	if a.cancel != nil {
		// Inline the stop sequence instead of calling Stop() to avoid
		// releasing the lock (which would allow concurrent Stop/Start to
		// observe inconsistent state). Nil-ing cancel under the lock means
		// any concurrent Stop() becomes a no-op.
		cancel := a.cancel
		a.cancel = nil
		a.mu.Unlock()
		slog.Warn("agent already running, stopping previous instance before restart")
		cancel()
		a.wg.Wait()
		a.mu.Lock()
	}
	ctx, a.cancel = context.WithCancel(ctx)

	a.wg.Add(4)
	a.mu.Unlock()

	go func() {
		defer a.wg.Done()
		a.Monitor(ctx)
	}()

	go func() {
		defer a.wg.Done()
		a.Prune(ctx)
	}()

	go func() {
		defer a.wg.Done()
		a.InvestigationScheduler(ctx)
	}()

	go func() {
		defer a.wg.Done()
		a.runPipelineSweeper(ctx)
	}()

	slog.Info("agent started, monitoring + pruner + scheduler + pipeline-sweeper goroutines spawned")
}

// Stop cancels the monitoring goroutine and waits for it to finish.
func (a *Agent) Stop() {
	a.mu.Lock()
	cancel := a.cancel
	a.cancel = nil
	a.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	a.wg.Wait()
	slog.Info("agent stopped")
}
