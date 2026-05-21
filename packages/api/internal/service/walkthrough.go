package service

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/walkthrough"
)

// WalkthroughService coordinates the walkthrough cache, the credential
// lookup, and the LLM call. Handlers feed it the resolved RepositoryState
// and the request parameters; the service returns the resulting record
// plus a flag indicating whether it came from cache.
type WalkthroughService struct {
	walkthroughs *storage.WalkthroughRepo
	preferences  *storage.PreferenceRepo
}

// ErrAnthropicNotConfigured signals that no API key was found in env vars
// or saved preferences; handlers translate to 412 Precondition Failed.

type GenerateRequest struct {
	State      review.RepositoryState
	Kind       string
	ContextSHA string
	Refresh    bool
	UserID     int64
}

type GenerateResult struct {
	Record storage.WalkthroughRecord
	Cached bool
}

func NewWalkthroughService(walkthroughs *storage.WalkthroughRepo, preferences *storage.PreferenceRepo) *WalkthroughService {
	return &WalkthroughService{walkthroughs: walkthroughs, preferences: preferences}
}

var ErrAnthropicNotConfigured = errors.New("anthropic api key not configured")

// Generate runs the cache lookup, falls through to the LLM when the cache
// is stale (or absent), and persists the new record. Caching failures are
// non-fatal: the caller still gets the fresh result.
func (s *WalkthroughService) Generate(
	ctx context.Context,
	req GenerateRequest,
) (GenerateResult, error) {
	cached, found, err := s.walkthroughs.GetWalkthrough(ctx, req.State.Root, req.Kind, req.ContextSHA)

	if err != nil {
		return GenerateResult{}, err
	}

	fingerprint := walkthrough.FingerprintForState(req.State)

	if !req.Refresh && found && cached.Fingerprint == fingerprint {
		return GenerateResult{Record: cached, Cached: true}, nil
	}

	apiKey, modelID := s.credentials(ctx, req.UserID)

	if apiKey == "" {
		return GenerateResult{}, ErrAnthropicNotConfigured
	}

	result, err := walkthrough.Generate(ctx, walkthrough.Request{
		APIKey:  apiKey,
		ModelID: modelID,
		State:   req.State,
	})

	if err != nil {
		return GenerateResult{}, err
	}

	record := storage.WalkthroughRecord{
		RepoRoot:    req.State.Root,
		ContextKind: req.Kind,
		ContextSHA:  req.ContextSHA,
		Fingerprint: result.Fingerprint,
		ModelID:     result.ModelID,
		Order:       result.Order,
		Notes:       result.Notes,
		Summary:     result.Summary,
		GeneratedAt: result.GeneratedAt,
	}

	_ = s.walkthroughs.UpsertWalkthrough(ctx, record)

	return GenerateResult{Record: record}, nil
}

// credentials sources the Anthropic API key and optional model override.
// Env vars win over the user's saved UI preferences so a developer can
// override without touching settings.
func (s *WalkthroughService) credentials(ctx context.Context, userID int64) (string, string) {
	if env := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); env != "" {
		return env, strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL"))
	}

	if userID == 0 {
		return "", ""
	}

	prefs, err := s.preferences.GetUIPreferences(ctx, userID)

	if err != nil {
		return "", ""
	}

	apiKey := strings.TrimSpace(prefs.Values[storage.PrefKeyAnthropicAPIKey])
	modelID := strings.TrimSpace(prefs.Values[storage.PrefKeyAnthropicModel])

	return apiKey, modelID
}
