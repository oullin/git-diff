package ai

import (
	"fmt"
	"sort"
	"sync"
)

// Registry is safe for concurrent use; Register/Get can race in tests
// and hot-reload scenarios.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register returns the previous provider for the same ID, if any, so
// tests can swap and restore.
func (r *Registry) Register(p Provider) Provider {
	r.mu.Lock()

	defer r.mu.Unlock()

	prev := r.providers[p.ID()]
	r.providers[p.ID()] = p

	return prev
}

// Get's error lists the available IDs so the HTTP layer can surface a
// useful 412 to the renderer.
func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	p, ok := r.providers[id]

	if !ok {
		return nil, fmt.Errorf("ai provider %q not registered; available: %v", id, r.idsLocked())
	}

	return p, nil
}

// List returns every registered provider ID, sorted.
func (r *Registry) List() []string {
	r.mu.RLock()

	defer r.mu.RUnlock()

	return r.idsLocked()
}

func (r *Registry) idsLocked() []string {
	ids := make([]string, 0, len(r.providers))

	for id := range r.providers {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}
