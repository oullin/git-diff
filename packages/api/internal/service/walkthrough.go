package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocanto/git-diff/internal/ai"
	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/userconfig"
	"github.com/gocanto/git-diff/internal/walkthrough"
)

// WalkthroughService coordinates the cache, the user config, the AI
// provider registry, and the walkthrough orchestrator. Handlers feed it
// the resolved RepositoryState and the request parameters; the service
// returns the resulting record plus a flag indicating cache hit.
//
// Depends on the registry/reader INTERFACES, not concretes — keeps the
// service layer testable without booting viper or hitting real LLMs.
type WalkthroughService struct {
	walkthroughs *storage.WalkthroughRepo
	providers    *ai.Registry
	userConfig   userconfig.Reader
}

// ErrProviderUnavailable signals that the configured provider couldn't
// be reached (missing API key, missing CLI binary). Handlers translate
// to 412 Precondition Failed so the renderer can surface a specific
// remediation prompt.

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

// NewWalkthroughService composes the dependencies. Wired in the app
// bootstrap; tests can pass fakes for any dependency individually.

// Generate runs the cache lookup, falls through to the LLM when the
// cache is stale (or absent), and persists the new record. Caching
// failures are non-fatal: the caller still gets the fresh result.

// providerWithModel wraps a Provider so its Generate call sees the
// caller's preferred Model on the request. The underlying provider is
// unchanged — we just stamp the Model field on the way through.

type modelOverrideProvider struct {
	ai.Provider
	model string
}

var ErrProviderUnavailable = errors.New("ai provider unavailable")

func NewWalkthroughService(
	walkthroughs *storage.WalkthroughRepo,
	providers *ai.Registry,
	userConfig userconfig.Reader,
) *WalkthroughService {
	return &WalkthroughService{
		walkthroughs: walkthroughs,
		providers:    providers,
		userConfig:   userConfig,
	}
}

func (s *WalkthroughService) Generate(
	ctx context.Context,
	req GenerateRequest,
) (GenerateResult, error) {
	cfg := s.userConfig.Get()

	provider, err := s.providers.Get(cfg.Walkthrough.Provider)

	if err != nil {
		return GenerateResult{}, fmt.Errorf("%w: %w", ErrProviderUnavailable, err)
	}

	fingerprint := walkthrough.FingerprintForStateAndProvider(req.State, provider.ID())

	cached, found, err := s.walkthroughs.GetWalkthrough(ctx, req.State.Root, req.Kind, req.ContextSHA)

	if err != nil {
		return GenerateResult{}, err
	}

	if !req.Refresh && found && cached.Fingerprint == fingerprint {
		return GenerateResult{Record: cached, Cached: true}, nil
	}

	result, err := walkthrough.Generate(ctx, walkthrough.Request{
		Provider: providerWithModel(provider, cfg.Walkthrough.Model),
		State:    req.State,
		Budget: walkthrough.Budget{
			PerFileBytes: cfg.Walkthrough.PerFileBudgetBytes,
			TotalBytes:   cfg.Walkthrough.PatchBudgetBytes,
		},
	})

	if err != nil {
		return GenerateResult{}, err
	}

	record := storage.WalkthroughRecord{
		RepoRoot:    req.State.Root,
		ContextKind: req.Kind,
		ContextSHA:  req.ContextSHA,
		Fingerprint: result.Fingerprint,
		ProviderID:  result.ProviderID,
		ModelID:     result.ModelID,
		Groups:      toStorageGroups(result.Groups),
		Order:       result.Order,
		Notes:       result.Notes,
		Summary:     result.Summary,
		GeneratedAt: result.GeneratedAt,
	}

	_ = s.walkthroughs.UpsertWalkthrough(ctx, record)

	return GenerateResult{Record: record}, nil
}

func providerWithModel(p ai.Provider, model string) ai.Provider {
	return modelOverrideProvider{Provider: p, model: model}
}

func (m modelOverrideProvider) Generate(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
	if req.Model == "" {
		req.Model = m.model
	}

	return m.Provider.Generate(ctx, req)
}

func toStorageGroups(groups []walkthrough.Group) []storage.WalkthroughGroup {
	out := make([]storage.WalkthroughGroup, 0, len(groups))

	for _, group := range groups {
		files := make([]storage.WalkthroughGroupFile, 0, len(group.Files))

		for _, file := range group.Files {
			files = append(files, storage.WalkthroughGroupFile{
				Path:   file.Path,
				Note:   file.Note,
				Action: string(file.Action),
				Impact: string(file.Impact),
			})
		}

		out = append(out, storage.WalkthroughGroup{
			ID:        group.ID,
			Title:     group.Title,
			Rationale: group.Rationale,
			Files:     files,
		})
	}

	return out
}
