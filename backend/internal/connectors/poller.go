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
type Poller struct {
	queries *db.Queries
	mu      sync.Mutex
	wg      sync.WaitGroup
	pollers map[uuid.UUID]pollerEntry
}

type pollerEntry struct {
	cancel context.CancelFunc
	done   chan struct{} // closed when the goroutine exits
}

// NewPoller creates a new Poller manager.
func NewPoller(queries *db.Queries) *Poller {
	return &Poller{
		queries: queries,
		pollers: make(map[uuid.UUID]pollerEntry),
	}
}

// Start launches a goroutine that calls conn.Poll() on the given interval.
// If a poller for this connectionID already exists, it is stopped first.
// Interval is clamped to a minimum of 5s to prevent ticker panics.
func (p *Poller) Start(conn PollConnector, connectionID uuid.UUID, interval time.Duration) {
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
		p.run(ctx, conn, connectionID, interval)
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

func (p *Poller) run(ctx context.Context, conn PollConnector, connectionID uuid.UUID, interval time.Duration) {
	// pollWithTimeout wraps each poll call with a per-poll deadline so a hung
	// upstream can't block the goroutine indefinitely.
	pollWithTimeout := func() {
		timeout := 2 * interval
		if timeout < 30*time.Second {
			timeout = 30 * time.Second
		}
		pollCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := conn.Poll(pollCtx, p.queries); err != nil {
			slog.Error("poll failed", "connection_id", connectionID, "err", err)
		}
	}

	// Run once immediately.
	pollWithTimeout()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pollWithTimeout()
		case <-ctx.Done():
			return
		}
	}
}
