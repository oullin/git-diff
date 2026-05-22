package httpx

import (
	"net/http"

	"github.com/gocanto/git-diff/internal/review"
)

func withRequestCaches(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := review.WithRootCache(r.Context())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
