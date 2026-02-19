package agent

import (
	"context"
	"log/slog"
)

type Agent struct {
	// TODO: add dependencies (db queries, connectors, memory client, anthropic client)
}

func New() *Agent {
	return &Agent{}
}

func (a *Agent) Start(ctx context.Context) {
	slog.Info("agent started")
	// TODO: start monitoring goroutine
}

func (a *Agent) Stop() {
	slog.Info("agent stopped")
}
