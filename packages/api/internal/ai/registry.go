package ai

import (
	"fmt"
	"sort"
	"sync"
)

// Registry holds the active set of providers and exposes lookup by ID.
// Empty until callers Register implementations — typically in app
// bootstrap, once per binary.
//
// Safe for concurrent use; Register and Get can race in tests / hot
// reload scenarios.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds (or replaces) a provider keyed by its ID. Returns the
// previous provider for that ID, if any, so tests can swap and restore.
func (r *Registry) Register(p Provider) Provider {
	r.mu.Lock()

	defer r.mu.Unlock()

	prev := r.providers[p.ID()]
	r.providers[p.ID()] = p

	return prev
}

// Get returns the provider for id, or an error listing the available
// ones when the id isn't registered. The error is shaped so the HTTP
// layer can surface a useful 412 to the renderer.
func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	p, ok := r.providers[id]

	if !ok {
		return nil, fmt.Errorf("ai provider %q not registered; available: %v", id, r.idsLocked())
	}

	return p, nil
}

// List returns every registered provider ID, sorted. Useful for /v1
// discovery handlers and config-validation error messages.
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
