package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/walkthrough"
)

type walkthroughRequest struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	SHA     string `json:"sha"`
	Refresh bool   `json:"refresh"`
}

type walkthroughResponse struct {
	storage.WalkthroughRecord
	Stale bool `json:"stale"`
}

func (s Server) walkthroughGenerate(w http.ResponseWriter, r *http.Request) {
	var req walkthroughRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	if req.Path == "" {
		req.Path = s.Repo
	}

	if req.Kind == "" {
		req.Kind = review.RepositoryModeWorking
	}

	state, err := loadStateForWalkthrough(r.Context(), req)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)

		return
	}

	userID := int64(0)

	if s.Auth != nil {
		userID = s.Auth.CurrentUserID()
	}

	result, err := s.Services.Walkthroughs().Generate(r.Context(), service.GenerateRequest{
		State:      state,
		Kind:       req.Kind,
		ContextSHA: req.SHA,
		Refresh:    req.Refresh,
		UserID:     userID,
	})

	if err != nil {
		switch {
		case errors.Is(err, service.ErrAnthropicNotConfigured),
			errors.Is(err, walkthrough.ErrMissingAPIKey):
			writeError(w, http.StatusPreconditionFailed, err)
		default:
			writeError(w, http.StatusBadGateway, err)
		}

		return
	}

	writeJSON(w, http.StatusOK, walkthroughResponse{
		WalkthroughRecord: result.Record,
		Stale:             false,
	})
}

func loadStateForWalkthrough(
	ctx context.Context,
	req walkthroughRequest,
) (review.RepositoryState, error) {
	if req.Kind == review.RepositoryModeCommit {
		if strings.TrimSpace(req.SHA) == "" {
			return review.RepositoryState{}, errors.New("sha is required for commit walkthroughs")
		}

		return review.ReadCommitState(ctx, req.Path, req.SHA)
	}

	return review.ReadRepositoryState(ctx, req.Path)
}
