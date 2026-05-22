package httpx

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicAuthPath(r.URL.Path) {
			next.ServeHTTP(w, r)

			return
		}

		if s.Session == nil || s.Session.CurrentUserID() == 0 {
			writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

			return
		}

		next.ServeHTTP(w, r)
	})
}

func isPublicAuthPath(path string) bool {
	if path == "/v1/healthz" {
		return true
	}

	return strings.HasPrefix(path, "/v1/auth/")
}
