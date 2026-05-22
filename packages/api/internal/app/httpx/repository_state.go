package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
)

func (s Server) repositoryState(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	state, err := review.ReadRepositoryState(r.Context(), path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryOpen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if s.Auth != nil && s.Auth.CurrentUserID() != 0 {
		userID := s.Auth.CurrentUserID()
		_, _ = s.Services.preferences.Save(r.Context(), userID, map[string]string{
			storage.PrefKeyLastRepoRoot: state.Root,
		})
		_, _ = s.Services.repos.Upsert(r.Context(), userID, state.Root, "")
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryRefresh(w http.ResponseWriter, r *http.Request) {
	s.repositoryOpen(w, r)
}

func (s Server) repositoryCommit(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	sha := r.URL.Query().Get("sha")

	if path == "" {
		path = s.Repo
	}

	if sha == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("sha is required"))

		return
	}

	state, err := review.ReadCommitState(r.Context(), path, sha)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryLog(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	var limit int

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)

		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid 'limit' parameter: %w", err))

			return
		}

		limit = parsed
	}

	commits, err := review.ListCommitLog(r.Context(), path, limit)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"commits": commits})
}

func (s Server) repositoryFile(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	if root == "" {
		root = s.Repo
	}

	if root == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("root is required"))

		return
	}

	if path == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("path is required"))

		return
	}

	file, err := review.ReadRepositoryFile(r.Context(), root, path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, file)
}

func (s Server) repositoryFileRange(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	root := query.Get("root")
	path := query.Get("path")
	ref := query.Get("ref")
	startParam := query.Get("startLine")
	endParam := query.Get("endLine")

	if root == "" {
		root = s.Repo
	}

	if root == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("root is required"))

		return
	}

	if path == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("path is required"))

		return
	}

	if startParam == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("startLine is required"))

		return
	}

	if endParam == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("endLine is required"))

		return
	}

	startLine, err := strconv.Atoi(startParam)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid 'startLine': %w", err))

		return
	}

	endLine, err := strconv.Atoi(endParam)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid 'endLine': %w", err))

		return
	}

	result, err := review.ReadRepositoryFileRange(r.Context(), root, path, ref, startLine, endLine)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, result)
}
