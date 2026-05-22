package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
)

func writeCollaboratorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAuthenticationRequired):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrRepositoryPathRequired):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, storage.ErrRepositoryNotOwned):
		writeError(w, http.StatusForbidden, err)
	case errors.Is(err, storage.ErrRepositoryNotFound):
		writeError(w, http.StatusNotFound, err)
	default:
		writeError(w, http.StatusBadRequest, err)
	}
}

func (s Server) listCollaborators(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		writeError(w, http.StatusBadRequest, errors.New("path query parameter is required"))

		return
	}

	collaborators, err := s.repos.ListCollaborators(
		r.Context(),
		s.Session.CurrentUserID(),
		path,
	)

	if err != nil {
		writeCollaboratorError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"collaborators": collaborators})
}

func (s Server) addCollaborator(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path   string `json:"path"`
		UserID int64  `json:"userId"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	collaborator, err := s.repos.GrantCollaborator(
		r.Context(),
		s.Session.CurrentUserID(),
		body.Path,
		body.UserID,
		body.Role,
	)

	if err != nil {
		writeCollaboratorError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, collaborator)
}

func (s Server) removeCollaborator(w http.ResponseWriter, r *http.Request) {
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

	if err := s.repos.RevokeCollaborator(
		r.Context(),
		s.Session.CurrentUserID(),
		path,
		targetID,
	); err != nil {
		writeCollaboratorError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
