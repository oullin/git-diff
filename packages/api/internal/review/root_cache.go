package review

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// rootCacheKey is the unexported context key under which a per-request git
// toplevel resolver is stored. Use WithRootCache to install one; RootFor to
// read through it.
type rootCacheKey struct{}

// rootCache memoises `git rev-parse --show-toplevel` for a single request.
// The cache is keyed by launch path because two different launch paths can
// legitimately resolve to the same root (one via symlink, one direct) — we
// don't want to entangle them.
type rootCache struct {
	mu      sync.Mutex
	entries map[string]rootCacheEntry
}

type rootCacheEntry struct {
	root string
	err  error
}

// WithRootCache returns a context that memoises git toplevel lookups for
// the lifetime of the request. Call once at the top of each handler.
//
// Without an installed cache, RootFor still works but falls through to git
// every time — that's by design so unit tests don't need to bootstrap the
// context plumbing.
func WithRootCache(ctx context.Context) context.Context {
	if existing, _ := ctx.Value(rootCacheKey{}).(*rootCache); existing != nil {
		return ctx
	}

	return context.WithValue(ctx, rootCacheKey{}, &rootCache{
		entries: make(map[string]rootCacheEntry),
	})
}

// RootFor resolves the git toplevel for launchPath, using the request-scoped
// cache when present. Equivalent to `git rev-parse --show-toplevel`.
//
// All review-package code that needs the repo root should go through this
// helper instead of calling gitOutput directly — that's how the cache earns
// its keep across the 7+ call sites that used to invoke rev-parse.
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
