package protocol

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(adapter Adapter) error {
	if adapter == nil { return fmt.Errorf("protocol: nil adapter") }
	d := adapter.Descriptor()
	if d.ID == "" || d.Version == "" || d.RuntimeFamily == "" {
		return fmt.Errorf("protocol: incomplete descriptor")
	}
	if d.Capabilities.Accounting == "" {
		return fmt.Errorf("protocol %s: accounting capability is required", d.ID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[d.ID]; exists {
		return fmt.Errorf("protocol %s: duplicate adapter", d.ID)
	}
	r.adapters[d.ID] = adapter
	return nil
}

func (r *Registry) Get(id string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[id]
	return adapter, ok
}
