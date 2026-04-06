package agent

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
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

type Agent struct {
	queries      *db.Queries
	client       *anthropic.Client
	config       *config.Config
	classifier   Classifier
	notifier     *notifications.Dispatcher
	githubClient *github.Client
	limiter      *rate.Limiter
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier, notifier *notifications.Dispatcher, gh *github.Client) *Agent {
	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))
	return &Agent{
		queries:      queries,
		client:       &client,
		config:       cfg,
		classifier:   classifier,
		notifier:     notifier,
		githubClient: gh,
		limiter:      rate.NewLimiter(monitorLLMRate, 5),
	}
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
