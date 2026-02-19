package connectors

import "sync"

// Registry manages all active connectors at runtime.
type Registry struct {
	mu         sync.RWMutex
	connectors map[string]Connector
}

func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]Connector),
	}
}

func (r *Registry) Register(id string, c Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors[id] = c
}

func (r *Registry) Get(id string) (Connector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.connectors[id]
	return c, ok
}

func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.connectors[id]
	if !ok {
		return nil
	}
	delete(r.connectors, id)
	return c.Close()
}

func (r *Registry) All() map[string]Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]Connector, len(r.connectors))
	for k, v := range r.connectors {
		out[k] = v
	}
	return out
}
