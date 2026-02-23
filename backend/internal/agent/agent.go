package agent

import (
	"context"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

type Agent struct {
	queries *db.Queries
	client  *anthropic.Client
	config  *config.Config
}

func New(queries *db.Queries, cfg *config.Config) *Agent {
	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))
	return &Agent{
		queries: queries,
		client:  &client,
		config:  cfg,
	}
}

func (a *Agent) Start(ctx context.Context) {
	slog.Info("agent started")
	// TODO: start monitoring goroutine
}

func (a *Agent) Stop() {
	slog.Info("agent stopped")
}
