package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/review"
)

func (s Server) repositoryBranches(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		path = s.Repo
	}

	result, err := s.Services.branches.List(r.Context(), path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	payload := map[string]any{"branches": result.Names}

	if result.Records != nil {
		payload["records"] = result.Records
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s Server) repositoryCheckout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if body.Path == "" {
		body.Path = s.Repo
	}

	if err := review.CheckoutBranch(r.Context(), body.Path, body.Branch); err != nil {
		var dirty *review.WorkingTreeDirtyError

		if errors.As(err, &dirty) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error": dirty.Error(),
				"code":  "working_tree_dirty",
				"files": dirty.Files,
			})

			return
		}

		writeError(w, http.StatusBadRequest, err)

		return
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s Server) repositoryCreateBranch(w http.ResponseWriter, r *http.Request) {
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

	if err := review.CreateBranch(r.Context(), body.Path, body.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	state, err := review.ReadRepositoryState(r.Context(), body.Path)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}
