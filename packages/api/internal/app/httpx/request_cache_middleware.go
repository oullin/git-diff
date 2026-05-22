package httpx

import (
	"net/http"

	"github.com/gocanto/git-diff/internal/review"
)

// withRequestCaches installs the per-request memoisation primitives owned
// by the review package — currently a git-toplevel cache. Wraps the
// downstream handler so every git.RootFor call within a single HTTP
// request shares its lookup.
//
// Kept as a standalone middleware (single responsibility: install caches)
// so adding a future cache only changes this one file.
func withRequestCaches(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := review.WithRootCache(r.Context())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
