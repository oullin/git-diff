package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gocanto/git-diff/internal/review"
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

	store, closeStore, err := s.Store(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	defer closeStore()

	cached, found, err := store.GetWalkthrough(r.Context(), state.Root, req.Kind, req.SHA)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)

		return
	}

	fingerprint := walkthrough.FingerprintForState(state)

	if !req.Refresh && found && cached.Fingerprint == fingerprint {
		writeJSON(w, http.StatusOK, walkthroughResponse{WalkthroughRecord: cached, Stale: false})

		return
	}

	userID := int64(0)

	if s.Auth != nil {
		userID = s.Auth.CurrentUserID()
	}

	apiKey, modelID := walkthroughCredentials(r.Context(), store, userID)

	if apiKey == "" {
		writeError(w, http.StatusPreconditionFailed, errors.New("anthropic api key not configured"))

		return
	}

	result, err := walkthrough.Generate(r.Context(), walkthrough.Request{
		APIKey:  apiKey,
		ModelID: modelID,
		State:   state,
	})

	if err != nil {
		if errors.Is(err, walkthrough.ErrMissingAPIKey) {
			writeError(w, http.StatusPreconditionFailed, err)

			return
		}

		writeError(w, http.StatusBadGateway, err)

		return
	}

	record := storage.WalkthroughRecord{
		RepoRoot:    state.Root,
		ContextKind: req.Kind,
		ContextSHA:  req.SHA,
		Fingerprint: result.Fingerprint,
		ModelID:     result.ModelID,
		Order:       result.Order,
		Notes:       result.Notes,
		Summary:     result.Summary,
		GeneratedAt: result.GeneratedAt,
	}

	// Caching failure is non-fatal; the caller still gets the fresh result.
	_ = store.UpsertWalkthrough(r.Context(), record)

	writeJSON(w, http.StatusOK, walkthroughResponse{WalkthroughRecord: record, Stale: false})
}

func loadStateForWalkthrough(ctx context.Context, req walkthroughRequest) (review.RepositoryState, error) {
	if req.Kind == review.RepositoryModeCommit {
		if strings.TrimSpace(req.SHA) == "" {
			return review.RepositoryState{}, errors.New("sha is required for commit walkthroughs")
		}

		return review.ReadCommitState(ctx, req.Path, req.SHA)
	}

	return review.ReadRepositoryState(ctx, req.Path)
}

// walkthroughCredentials sources the Anthropic API key and optional model
// override. Env vars win over the user's saved UI preferences so a developer
// can override without touching settings.
func walkthroughCredentials(ctx context.Context, store *storage.Store, userID int64) (string, string) {
	if env := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); env != "" {
		return env, strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL"))
	}

	if userID == 0 {
		return "", ""
	}

	prefs, err := store.GetUIPreferences(ctx, userID)

	if err != nil {
		return "", ""
	}

	apiKey := strings.TrimSpace(prefs.Values[storage.PrefKeyAnthropicAPIKey])
	modelID := strings.TrimSpace(prefs.Values[storage.PrefKeyAnthropicModel])

	return apiKey, modelID
}
