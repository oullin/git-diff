package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/review"
	"github.com/oullin/git-diff/internal/service"
	"github.com/oullin/git-diff/internal/storage"
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

	if s.Session != nil {
		userID = s.Session.CurrentUserID()
	}

	result, err := s.walkthroughs.Generate(r.Context(), service.GenerateRequest{
		State:      state,
		Kind:       req.Kind,
		ContextSHA: req.SHA,
		Refresh:    req.Refresh,
		UserID:     userID,
	})

	if err != nil {
		switch {
		case errors.Is(err, service.ErrProviderUnavailable),
			errors.Is(err, ai.ErrCodexNotInstalled):
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
