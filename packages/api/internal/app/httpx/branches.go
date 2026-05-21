package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/service"
)

func (s Server) handleBranchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrBranchNameRequired):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrBranchAlreadyLocked):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusBadRequest, err)
	}
}

func (s Server) lockBranch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	branches, err := s.Services.Branches().Lock(r.Context(), s.Auth.CurrentUserID(), body.Path, body.Name)

	if err != nil {
		s.handleBranchError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (s Server) unlockBranch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	branches, err := s.Services.Branches().Unlock(r.Context(), s.Auth.CurrentUserID(), body.Path, body.Name)

	if err != nil {
		s.handleBranchError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (s Server) deleteBranch(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	err := s.Services.Branches().Delete(
		r.Context(),
		s.Auth.CurrentUserID(),
		path,
		r.URL.Query().Get("name"),
	)

	if err != nil {
		s.handleBranchError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
