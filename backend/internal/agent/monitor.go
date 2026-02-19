package agent

import (
	"context"
	"log/slog"
)

// Monitor runs the 24/7 monitoring goroutine that processes incoming
// log streams and triggers the agent when anomalies are detected.
func (a *Agent) Monitor(ctx context.Context) {
	slog.Info("monitoring goroutine started")

	// TODO: implement monitoring loop
	// - Ingest logs from connectors
	// - Detect anomalies
	// - Trigger agent loop when warranted

	<-ctx.Done()
	slog.Info("monitoring goroutine stopped")
}
