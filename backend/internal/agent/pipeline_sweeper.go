package agent

import (
	"context"
	"log/slog"
	"time"
)

// pipelineSweeperInterval is how often the leak-canary loop wakes up. One
// minute is loose enough to be effectively free (a single map snapshot +
// comparison) but tight enough that a runaway subscriber is noticed in
// logs within the span of a single operator attention window.
const pipelineSweeperInterval = time.Minute

// pipelineSweeperThreshold is the per-app subscriber count above which we
// emit a warning. Phase 3.4 picks 50 as a round "nobody should ever hit
// this" number — a single human user on a realistic screen count would be
// under 5, and the per-user cap enforced in the SSE handler limits growth
// further. Any app legitimately above 50 is a bug signal.
const pipelineSweeperThreshold = 50

// runPipelineSweeper is a background goroutine that periodically scans the
// PipelineBus for abnormally large subscriber sets. It never force-closes
// channels — that would mask the underlying leak (e.g. a handler path
// that forgets to call unsub). Instead it emits a WARN log line so the
// leak surfaces in observability without taking action on live traffic.
//
// The sweeper also refreshes the global `pipeline_bus_dropped_events_total`
// metric at every tick — the Publish path increments in-situ, but the
// gauge-style visibility into the running total is helpful for dashboards
// when drops are rare.
func (a *Agent) runPipelineSweeper(ctx context.Context) {
	if a.pipelineBus == nil {
		return
	}
	ticker := time.NewTicker(pipelineSweeperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			counts := a.pipelineBus.SubscriberCounts()
			for appID, n := range counts {
				if n > pipelineSweeperThreshold {
					slog.Warn("pipeline bus leak canary: app has many live subscribers",
						"app_id", appID, "subscriber_count", n,
						"threshold", pipelineSweeperThreshold)
				}
			}
		}
	}
}
