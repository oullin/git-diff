package review

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type rootCacheKey struct{}

// rootCache memoises `git rev-parse --show-toplevel` for a single request,
// keyed by launch path so symlinked paths don't collapse onto each other.
type rootCache struct {
	mu      sync.Mutex
	entries map[string]rootCacheEntry
}

type rootCacheEntry struct {
	root string
	err  error
}

// WithRootCache memoises git toplevel lookups for the request's lifetime.
// RootFor still works without one (it just hits git every time) so unit
// tests don't need to bootstrap the context plumbing.
func WithRootCache(ctx context.Context) context.Context {
	if existing, _ := ctx.Value(rootCacheKey{}).(*rootCache); existing != nil {
		return ctx
	}

	return context.WithValue(ctx, rootCacheKey{}, &rootCache{
		entries: make(map[string]rootCacheEntry),
	})
}

// RootFor resolves the git toplevel for launchPath through the
// request-scoped cache when present. All review-package code that needs
// the repo root should go through this helper rather than gitOutput
// directly — that's how the cache earns its keep.
func RootFor(ctx context.Context, launchPath string) (string, error) {
	cache, _ := ctx.Value(rootCacheKey{}).(*rootCache)

	if cache == nil {
		return resolveTopLevel(ctx, launchPath)
	}

	cache.mu.Lock()
	entry, ok := cache.entries[launchPath]
	cache.mu.Unlock()

	if ok {
		return entry.root, entry.err
	}

	root, err := resolveTopLevel(ctx, launchPath)

	cache.mu.Lock()
	cache.entries[launchPath] = rootCacheEntry{root: root, err: err}
	cache.mu.Unlock()

	return root, err
}

func resolveTopLevel(ctx context.Context, launchPath string) (string, error) {
	out, err := gitOutput(ctx, launchPath, "rev-parse", "--show-toplevel")

	if err != nil {
		return "", fmt.Errorf("resolve git root: %w", err)
	}

	return strings.TrimSpace(out), nil
}
