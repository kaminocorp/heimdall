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
	done     chan struct{} // closed when the goroutine exits
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
	// Extract the old entry (if any) under the lock, then release before
	// blocking on Close(). This prevents holding the mutex during the
	// potentially slow drain timeout.
	old, hadOld := m.listeners[connectionID]
	if hadOld {
		delete(m.listeners, connectionID)
	}
	m.mu.Unlock()

	if hadOld {
		old.cancel()
		old.listener.Close()
		<-old.done // wait for old goroutine to fully exit
	}

	lCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	m.mu.Lock()
	m.listeners[connectionID] = listenerEntry{listener: l, cancel: cancel, done: done}
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer close(done)
		if err := l.Listen(lCtx); err != nil {
			slog.Error("listener exited with error", "connection_id", connectionID, "err", err)
		}
	}()

	slog.Info("listener started", "connection_id", connectionID)
	return nil
}

// Stop shuts down the listener for the given connection and waits for it to exit.
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
		<-entry.done
		slog.Info("listener stopped", "connection_id", connectionID)
	}
}

// StopAll shuts down all active listeners and waits for them to finish.
// Unlike Stop() which waits on individual done channels, StopAll() uses
// wg.Wait() which is sufficient: each listener goroutine calls wg.Done()
// as its final action before the done channel would be signalled, so
// wg.Wait() returning guarantees all goroutines have completed.
func (m *ListenerManager) StopAll() {
	m.mu.Lock()
	entries := make(map[uuid.UUID]listenerEntry, len(m.listeners))
	for id, entry := range m.listeners {
		entries[id] = entry
	}
	m.listeners = make(map[uuid.UUID]listenerEntry)
	m.mu.Unlock()

	for id, entry := range entries {
		entry.cancel()
		entry.listener.Close()
		slog.Info("listener stopped", "connection_id", id)
	}

	m.wg.Wait()
}
