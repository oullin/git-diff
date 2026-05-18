package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/review"
)

func (s Server) lockBranch(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

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

	root, err := review.ResolveRoot(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	if err := store.LockBranch(r.Context(), root, body.Name, userID); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	branches, err := store.ListBranches(r.Context(), root)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (s Server) unlockBranch(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

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

	root, err := review.ResolveRoot(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	if err := store.UnlockBranch(r.Context(), root, body.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	branches, err := store.ListBranches(r.Context(), root)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (s Server) deleteBranch(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	path := r.URL.Query().Get("path")
	name := r.URL.Query().Get("name")

	if name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name query parameter is required"))

		return
	}

	if path == "" {
		path = s.Repo
	}

	root, err := review.ResolveRoot(r.Context(), path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	locked, err := store.IsBranchLocked(r.Context(), root, name)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	if locked {
		writeError(w, http.StatusConflict, errors.New("branch is locked"))

		return
	}

	if err := review.DeleteBranch(r.Context(), path, name); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	_ = store.DeleteBranchRow(r.Context(), root, name)

	w.WriteHeader(http.StatusNoContent)
}
