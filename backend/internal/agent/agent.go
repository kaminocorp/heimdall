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
	limiter      *rate.Limiter
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier, notifier *notifications.Dispatcher, gh *github.Client) *Agent {
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
		limiter:      rate.NewLimiter(monitorLLMRate, 5),
	}
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

// Start launches the monitoring goroutine.
// Safe to call multiple times: stops the previous instance first.
func (a *Agent) Start(ctx context.Context) {
	if a.cancel != nil {
		slog.Warn("agent already running, stopping previous instance before restart")
		a.Stop()
	}
	ctx, a.cancel = context.WithCancel(ctx)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.Monitor(ctx)
	}()
	slog.Info("agent started, monitoring goroutine spawned")
}

// Stop cancels the monitoring goroutine and waits for it to finish.
func (a *Agent) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
	slog.Info("agent stopped")
}
