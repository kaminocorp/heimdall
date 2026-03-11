package agent

import (
	"context"
	"log/slog"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/notifications"
)

type Agent struct {
	queries    *db.Queries
	client     *anthropic.Client
	config     *config.Config
	classifier Classifier
	notifier   *notifications.Dispatcher
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier, notifier *notifications.Dispatcher) *Agent {
	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))
	return &Agent{
		queries:    queries,
		client:     &client,
		config:     cfg,
		classifier: classifier,
		notifier:   notifier,
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
