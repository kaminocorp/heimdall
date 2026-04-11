package agent

import (
	"context"
	"log/slog"
	"time"
)

// pruneInterval is how often the pruner wakes up to delete expired rows from
// log_buffer. The retention window itself (48h) is defined in the SQL query
// PruneExpiredLogs — not here — so the DB is the single source of truth.
const pruneInterval = 1 * time.Hour

// Prune runs a background loop that periodically deletes expired rows from
// log_buffer. It runs once immediately on startup so a restart loop can't
// indefinitely delay the first prune, and then ticks every pruneInterval.
//
// The loop exits cleanly when ctx is cancelled. Errors from the underlying
// query are logged and swallowed — the pruner is best-effort and must never
// take down the agent.
func (a *Agent) Prune(ctx context.Context) {
	slog.Info("pruner goroutine started", "interval", pruneInterval)
	defer slog.Info("pruner goroutine stopped")

	ticker := time.NewTicker(pruneInterval)
	defer ticker.Stop()

	// Run once on startup so restarts don't delay the first prune by up to
	// a full interval. This also means the very first deploy starts shedding
	// old rows immediately rather than an hour later.
	a.pruneTick(ctx)

	for {
		select {
		case <-ticker.C:
			a.pruneTick(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// pruneTick performs a single PruneExpiredLogs call, logging success or
// failure. Exposed as a method (rather than inlined in Prune) so tests can
// exercise it directly without waiting on the ticker.
func (a *Agent) pruneTick(ctx context.Context) {
	rows, err := a.queries.PruneExpiredLogs(ctx)
	if err != nil {
		slog.Error("pruner: failed to delete expired logs", "err", err)
		return
	}
	if rows > 0 {
		slog.Info("pruner: deleted expired logs", "rows", rows)
	}
}
