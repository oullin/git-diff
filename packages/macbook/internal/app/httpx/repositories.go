package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (s Server) listRepositories(w http.ResponseWriter, r *http.Request) {
	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	repos, err := store.ListRepositories(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

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

	store, closeStore, err := s.WorkflowStore(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	repo, err := store.UpsertRepository(r.Context(), body.Path, body.Name)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, repo)
}

func (s Server) removeRepository(w http.ResponseWriter, r *http.Request) {
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

	if err := store.RemoveRepository(r.Context(), path); err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
