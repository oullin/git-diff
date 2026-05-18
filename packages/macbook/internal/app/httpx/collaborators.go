package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) listCollaborators(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	path := r.URL.Query().Get("path")

	if path == "" {
		writeError(w, http.StatusBadRequest, errors.New("path query parameter is required"))

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	collaborators, err := store.ListRepositoryCollaborators(r.Context(), userID, path)

	if err != nil {
		writeCollaboratorError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"collaborators": collaborators})
}

func (s Server) addCollaborator(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	var body struct {
		Path   string `json:"path"`
		UserID int64  `json:"userId"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	collaborator, err := store.GrantRepositoryAccess(r.Context(), userID, body.Path, body.UserID, body.Role)

	if err != nil {
		writeCollaboratorError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, collaborator)
}

func (s Server) removeCollaborator(w http.ResponseWriter, r *http.Request) {
	userID := s.Auth.CurrentUserID()

	if userID == 0 {
		writeError(w, http.StatusUnauthorized, errors.New("authentication required"))

		return
	}

	path := r.URL.Query().Get("path")
	target := r.URL.Query().Get("userId")

	if path == "" || target == "" {
		writeError(w, http.StatusBadRequest, errors.New("path and userId query parameters are required"))

		return
	}

	targetID, err := strconv.ParseInt(target, 10, 64)

	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("userId must be an integer"))

		return
	}

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	if err := store.RevokeRepositoryAccess(r.Context(), userID, path, targetID); err != nil {
		writeCollaboratorError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeCollaboratorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrRepositoryNotOwned):
		writeError(w, http.StatusForbidden, err)
	case errors.Is(err, storage.ErrRepositoryNotFound):
		writeError(w, http.StatusNotFound, err)
	default:
		writeError(w, http.StatusBadRequest, err)
	}
}
