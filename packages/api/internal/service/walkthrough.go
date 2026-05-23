package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocanto/git-diff/internal/ai"
	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/usercfg"
	"github.com/gocanto/git-diff/internal/walks"
)

type WalkthroughService struct {
	walkthroughs *storage.WalkthroughRepo
	providers    *ai.Registry
	userConfig   usercfg.Reader
}

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

type modelOverrideProvider struct {
	ai.Provider
	model string
}

var ErrProviderUnavailable = errors.New("ai provider unavailable")

func NewWalkthroughService(
	walkthroughs *storage.WalkthroughRepo,
	providers *ai.Registry,
	userConfig usercfg.Reader,
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

	fingerprint := walks.FingerprintForStateAndProvider(req.State, provider.ID())

	cached, found, err := s.walkthroughs.GetWalkthrough(ctx, req.State.Root, req.Kind, req.ContextSHA)

	if err != nil {
		return GenerateResult{}, err
	}

	if !req.Refresh && found && cached.Fingerprint == fingerprint {
		return GenerateResult{Record: cached, Cached: true}, nil
	}

	result, err := walks.Generate(ctx, walks.Request{
		Provider: providerWithModel(provider, cfg.Walkthrough.Model),
		State:    req.State,
		Budget: walks.Budget{
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

func toStorageGroups(groups []walks.Group) []storage.WalkthroughGroup {
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
