package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oullin/git-diff/internal/service"
	"github.com/oullin/git-diff/internal/storage"
)

func (s Server) handleRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrRepositoryPathRequired):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, storage.ErrRepositoryNotOwned):
		writeError(w, http.StatusForbidden, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func (s Server) listRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := s.repos.List(r.Context(), s.Session.CurrentUserID())

	if err != nil {
		s.handleRepositoryError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"repositories": repos})
}

func (s Server) upsertRepository(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	repo, err := s.repos.Upsert(
		r.Context(),
		s.Session.CurrentUserID(),
		body.Path,
		body.Name,
	)

	if err != nil {
		if errors.Is(err, service.ErrAuthenticationRequired) {
			writeError(w, http.StatusUnauthorized, err)

			return
		}

		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, repo)
}

func (s Server) removeRepository(w http.ResponseWriter, r *http.Request) {
	err := s.repos.Remove(
		r.Context(),
		s.Session.CurrentUserID(),
		r.URL.Query().Get("path"),
	)

	if err != nil {
		s.handleRepositoryError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
