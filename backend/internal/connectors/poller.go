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
	pollers map[uuid.UUID]context.CancelFunc
}

// NewPoller creates a new Poller manager.
func NewPoller(queries *db.Queries) *Poller {
	return &Poller{
		queries: queries,
		pollers: make(map[uuid.UUID]context.CancelFunc),
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
	defer p.mu.Unlock()

	// Stop existing poller for this connection if running.
	if cancel, ok := p.pollers[connectionID]; ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.pollers[connectionID] = cancel

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.run(ctx, conn, connectionID, interval)
	}()
	slog.Info("poller started", "connection_id", connectionID, "interval", interval)
}

// Stop cancels the polling goroutine for the given connection.
func (p *Poller) Stop(connectionID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cancel, ok := p.pollers[connectionID]; ok {
		cancel()
		delete(p.pollers, connectionID)
		slog.Info("poller stopped", "connection_id", connectionID)
	}
}

// StopAll cancels all active polling goroutines and waits for them to finish.
// Called on server shutdown to ensure in-flight polls complete cleanly.
func (p *Poller) StopAll() {
	p.mu.Lock()
	for id, cancel := range p.pollers {
		cancel()
		slog.Info("poller stopped", "connection_id", id)
	}
	p.pollers = make(map[uuid.UUID]context.CancelFunc)
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
