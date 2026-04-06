package connectors

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Listener is a long-running network listener (e.g. syslog TCP/TLS).
// It is started via Connect + Listen and stopped via Close.
type Listener interface {
	Connector
	Listen(ctx context.Context) error
}

// ListenerManager manages long-running network listeners (one per connection).
// Analogous to Poller but for persistent listener goroutines rather than timer-driven polls.
type ListenerManager struct {
	mu        sync.Mutex
	wg        sync.WaitGroup
	listeners map[uuid.UUID]listenerEntry
}

type listenerEntry struct {
	listener Listener
	cancel   context.CancelFunc
}

// NewListenerManager creates a new ListenerManager.
func NewListenerManager() *ListenerManager {
	return &ListenerManager{
		listeners: make(map[uuid.UUID]listenerEntry),
	}
}

// Start connects the listener, then launches a goroutine that calls Listen.
// If a listener for this connectionID already exists, it is stopped first.
func (m *ListenerManager) Start(ctx context.Context, l Listener, connectionID uuid.UUID) error {
	// Connect (bind the port) synchronously so callers get immediate feedback.
	if err := l.Connect(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing listener if running.
	if entry, ok := m.listeners[connectionID]; ok {
		entry.cancel()
		entry.listener.Close()
	}

	lCtx, cancel := context.WithCancel(context.Background())
	m.listeners[connectionID] = listenerEntry{listener: l, cancel: cancel}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := l.Listen(lCtx); err != nil {
			slog.Error("listener exited with error", "connection_id", connectionID, "err", err)
		}
	}()

	slog.Info("listener started", "connection_id", connectionID)
	return nil
}

// Stop shuts down the listener for the given connection.
func (m *ListenerManager) Stop(connectionID uuid.UUID) {
	m.mu.Lock()
	entry, ok := m.listeners[connectionID]
	if ok {
		delete(m.listeners, connectionID)
	}
	m.mu.Unlock()

	if ok {
		entry.cancel()
		entry.listener.Close()
		slog.Info("listener stopped", "connection_id", connectionID)
	}
}

// StopAll shuts down all active listeners and waits for them to finish.
func (m *ListenerManager) StopAll() {
	m.mu.Lock()
	for id, entry := range m.listeners {
		entry.cancel()
		entry.listener.Close()
		slog.Info("listener stopped", "connection_id", id)
	}
	m.listeners = make(map[uuid.UUID]listenerEntry)
	m.mu.Unlock()

	m.wg.Wait()
}
