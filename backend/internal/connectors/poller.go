package connectors

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const minPollInterval = 5 * time.Second

// Poller manages one goroutine per active poll-based connection.
//
// Per-tick handoff: each call to a connector's Poll method runs inside a
// fresh UserQueries(ownerUserID) scope minted from pools.App. The
// connector is handed the resulting *db.Queries; every read/write it
// performs against Heimdall's DB sees `app.current_user_id ==
// ownerUserID` and so satisfies the post-Phase-7 RLS policies on
// log_buffer, connection_sources, etc.
//
// The transaction is per-tick rather than per-poller-lifetime because
// the poll work is short (a single ingestion batch) and we don't want
// long-held connections during the inter-tick sleep. Per-tick keeps the
// pool free and avoids the idle-in-txn surface.
type Poller struct {
	pools   *db.Pools
	mu      sync.Mutex
	wg      sync.WaitGroup
	pollers map[uuid.UUID]pollerEntry
}

type pollerEntry struct {
	cancel context.CancelFunc
	done   chan struct{} // closed when the goroutine exits
}

// NewPoller creates a new Poller manager. Holding *db.Pools (rather than
// *db.Queries) lets every per-tick poll mint its own UserQueries scope
// for the connection owner.
func NewPoller(pools *db.Pools) *Poller {
	return &Poller{
		pools:   pools,
		pollers: make(map[uuid.UUID]pollerEntry),
	}
}

// Start launches a goroutine that calls conn.Poll() on the given interval.
// If a poller for this connectionID already exists, it is stopped first.
// Interval is clamped to a minimum of 5s to prevent ticker panics.
//
// userID is the connection owner — used to scope the per-tick
// UserQueries transaction so RLS policies evaluate against the right
// caller.
func (p *Poller) Start(conn PollConnector, connectionID, userID uuid.UUID, interval time.Duration) {
	if interval < minPollInterval {
		interval = minPollInterval
	}

	p.mu.Lock()
	// Stop existing poller for this connection and wait for it to exit
	// before starting a new one, preventing concurrent polls.
	if entry, ok := p.pollers[connectionID]; ok {
		entry.cancel()
		p.mu.Unlock()
		<-entry.done // wait for old goroutine to finish
		p.mu.Lock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	p.pollers[connectionID] = pollerEntry{cancel: cancel, done: done}
	p.mu.Unlock()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer close(done)
		p.run(ctx, conn, connectionID, userID, interval)
	}()
	slog.Info("poller started", "connection_id", connectionID, "interval", interval)
}

// Stop cancels the polling goroutine for the given connection and waits
// for it to exit, preventing a subsequent Start from racing with the
// still-running old goroutine.
func (p *Poller) Stop(connectionID uuid.UUID) {
	p.mu.Lock()
	entry, ok := p.pollers[connectionID]
	if ok {
		entry.cancel()
		delete(p.pollers, connectionID)
	}
	p.mu.Unlock()

	if ok {
		<-entry.done
		slog.Info("poller stopped", "connection_id", connectionID)
	}
}

// StopAll cancels all active polling goroutines and waits for them to finish.
// Called on server shutdown to ensure in-flight polls complete cleanly.
func (p *Poller) StopAll() {
	p.mu.Lock()
	for id, entry := range p.pollers {
		entry.cancel()
		slog.Info("poller stopped", "connection_id", id)
	}
	p.pollers = make(map[uuid.UUID]pollerEntry)
	p.mu.Unlock()

	// Wait outside the lock so goroutines can finish without deadlocking.
	p.wg.Wait()
}

func (p *Poller) run(ctx context.Context, conn PollConnector, connectionID, userID uuid.UUID, interval time.Duration) {
	// pollWithTimeout wraps each poll call with a per-poll deadline so a hung
	// upstream can't block the goroutine indefinitely. The poll executes
	// inside a UserQueries(userID) transaction so connector writes see
	// the owner's RLS context.
	pollWithTimeout := func() {
		timeout := 2 * interval
		if timeout < 30*time.Second {
			timeout = 30 * time.Second
		}
		pollCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		// Tests pass nil pools (and nil userID) to exercise the
		// goroutine lifecycle without DB plumbing — fall through to a
		// queries-less Poll so mocks observe Poll calls without
		// dereferencing nil pools. Production main.go always wires
		// non-nil pools.
		if p.pools == nil {
			if err := conn.Poll(pollCtx, nil); err != nil {
				slog.Error("poll failed", "connection_id", connectionID, "err", err)
			}
			return
		}
		if err := p.pools.WithUserQueries(pollCtx, userID, func(q *db.Queries) error {
			return conn.Poll(pollCtx, q)
		}); err != nil {
			slog.Error("poll failed", "connection_id", connectionID, "err", err)
		}
	}

	// Run once immediately.
	pollWithTimeout()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pollWithTimeout()
		}
	}
}
