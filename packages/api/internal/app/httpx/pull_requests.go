package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gocanto/git-diff/internal/review"
)

func (s Server) repositoryPullRequests(w http.ResponseWriter, r *http.Request) {
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

	prs, err := review.ListPullRequests(r.Context(), path, limit)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, review.ErrGhUnavailable) {
			status = http.StatusPreconditionFailed
		}

		writeError(w, status, err)

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pullRequests": prs})
}

func (s Server) repositoryPullRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	numberRaw := r.URL.Query().Get("number")

	if path == "" {
		path = s.Repo
	}

	number, err := strconv.Atoi(numberRaw)

	if err != nil || number <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid number: %q", numberRaw))

		return
	}

	state, err := review.ReadPullRequestState(r.Context(), path, number)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, review.ErrGhUnavailable) {
			status = http.StatusPreconditionFailed
		}

		writeError(w, status, err)

		return
	}

	writeJSON(w, http.StatusOK, state)
}
