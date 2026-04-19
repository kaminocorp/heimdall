package agent

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/metrics"
)

// PipelineEvent is a single stage observation in the log-pipeline journey.
// Mirrors the columns of log_pipeline_events, with `Valid*` flags on the
// nullable fields so subscribers can tell an unset value from a zero value
// (e.g. confidence of 0.0 vs. "classified didn't run yet").
type PipelineEvent struct {
	ID           uuid.UUID
	LogID        uuid.UUID
	AppID        uuid.UUID
	Stage        string
	OccurredAt   time.Time

	SourceType string
	Severity   string

	Type       string
	Category   string
	Confidence float64
	Summary    string

	HasGate   bool
	Escalated bool
	RuleHit   string

	AssessmentID *uuid.UUID
}

// pipelineBusBufferSize is the per-subscriber channel capacity. Publishing is
// non-blocking: if a subscriber's channel is full the event is dropped for
// that subscriber only and the drop counter ticks up. 128 is comfortably
// above a monitoring-tick's burst (logBatchLimit=200 can produce up to ~200
// classified + 200 gate + up to 50 assessment events in a tick, but they
// land in order and a subscriber draining at a few kHz keeps up trivially).
const pipelineBusBufferSize = 128

// PipelineBus is the in-memory fan-out hub for live PipelineEvents. Events
// are scoped by app_id: subscribers register interest in a single app, and
// Publish routes each event only to that app's subscribers.
//
// Durability lives elsewhere — every event that reaches Publish should also
// be written to log_pipeline_events by the caller (see pipeline_writer.go).
// Losing a live event is recoverable via the /stream endpoint's replay
// query; losing a persisted row is not, so persistence runs first.
type PipelineBus struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan PipelineEvent]struct{}

	// dropped counts non-blocking sends that lost events to a full
	// subscriber channel. Logged + exported as a leak canary (Phase 3.1).
	dropped atomic.Uint64
}

// NewPipelineBus constructs an empty bus. Safe for concurrent use.
func NewPipelineBus() *PipelineBus {
	return &PipelineBus{
		subs: make(map[uuid.UUID]map[chan PipelineEvent]struct{}),
	}
}

// Subscribe returns a receive-only channel of events for a single app plus
// an unsubscribe func the caller MUST invoke when done (on defer in the SSE
// handler). Forgetting to unsubscribe would leak the channel and pin
// Publish's RLock hot path behind an ever-growing subscriber set.
func (b *PipelineBus) Subscribe(appID uuid.UUID) (<-chan PipelineEvent, func()) {
	ch := make(chan PipelineEvent, pipelineBusBufferSize)

	b.mu.Lock()
	if b.subs[appID] == nil {
		b.subs[appID] = make(map[chan PipelineEvent]struct{})
	}
	b.subs[appID][ch] = struct{}{}
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		if set, ok := b.subs[appID]; ok {
			if _, ok := set[ch]; ok {
				delete(set, ch)
				close(ch)
			}
			if len(set) == 0 {
				delete(b.subs, appID)
			}
		}
		b.mu.Unlock()
	}

	return ch, unsub
}

// Publish fans the event out to every subscriber of evt.AppID. Sends are
// non-blocking — a slow subscriber loses its copy but never back-pressures
// the producer. That's the right trade-off here: a full channel means the
// SSE consumer has been stalled long enough that any rolling window of
// interest has already expired; the client will reconnect and refetch via
// /bootstrap on the next heartbeat.
//
// We hold the RLock across the non-blocking sends rather than snapshotting
// the channel list and releasing first. The snapshot-then-send approach
// has a subtle race: a concurrent unsub closes the channel between the
// release and the send, panicking with "send on closed channel". Holding
// RLock during select+default is bounded by a single non-blocking check
// per subscriber, so the lock is released within microseconds even under
// pathological fan-out. Writers (Subscribe/unsub) wait briefly — no
// starvation risk because Publish never blocks.
func (b *PipelineBus) Publish(evt PipelineEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs[evt.AppID] {
		select {
		case ch <- evt:
		default:
			n := b.dropped.Add(1)
			metrics.PipelineBusDroppedEvents.Inc()
			// Log the first drop and then every power of two after that
			// to keep the signal loud without spamming. `app_id` is the
			// scope that matters — a leak localised to one app tells us
			// which tab to look at.
			if n == 1 || n&(n-1) == 0 {
				slog.Warn("pipeline bus: dropped event for slow subscriber",
					"app_id", evt.AppID, "stage", evt.Stage, "total_dropped", n)
			}
		}
	}
}

// Dropped returns the running count of events dropped due to full
// subscriber channels. Exposed for test assertions and the Phase 3
// `pipeline_bus_dropped_events_total` metric.
func (b *PipelineBus) Dropped() uint64 {
	return b.dropped.Load()
}

// SubscriberCount returns how many channels are currently subscribed to an
// app. Used by the leak-canary sweeper (Phase 3.4) to surface stuck
// subscribers.
func (b *PipelineBus) SubscriberCount(appID uuid.UUID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs[appID])
}

// SubscriberCounts returns a snapshot of subscriber counts per app. The
// returned map is owned by the caller — safe to mutate or hold. Used by
// the leak-canary sweeper to warn on apps with abnormally many live
// channels (which would indicate forgotten unsub paths).
func (b *PipelineBus) SubscriberCounts() map[uuid.UUID]int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(map[uuid.UUID]int, len(b.subs))
	for appID, set := range b.subs {
		out[appID] = len(set)
	}
	return out
}
