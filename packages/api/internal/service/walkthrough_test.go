package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/domain/repostate"
	"github.com/oullin/git-diff/internal/storage"
	"github.com/oullin/git-diff/internal/usercfg"
	"github.com/oullin/git-diff/internal/walks"
)

// walksFingerprint mirrors the service's internal fingerprint derivation so
// tests can pre-seed the cache with the value the service will look up.

// stubProvider is a deterministic ai.Provider used to drive
// WalkthroughService.Generate without hitting a network.
type stubProvider struct {
	id       string
	response ai.GenerateResponse
	err      error
}

func walksFingerprint(state repostate.RepositoryState, providerID string) string {
	return walks.FingerprintForStateAndProvider(state, providerID)
}

func (p *stubProvider) ID() string { return p.id }

func (p *stubProvider) DefaultModel() string { return "stub-model" }

func (p *stubProvider) SupportsModel(string) bool { return true }

func (p *stubProvider) Generate(ctx context.Context, _ ai.GenerateRequest) (ai.GenerateResponse, error) {
	if p.err != nil {
		return ai.GenerateResponse{}, p.err
	}

	return p.response, nil
}

func newWalkthroughService(t *testing.T, store *storage.Store, providers *ai.Registry, reader usercfg.Reader) *WalkthroughService {
	t.Helper()

	return NewWalkthroughService(store.Walkthroughs, providers, reader)
}

func TestWalkthroughGenerateProviderUnavailable(t *testing.T) {
	store := newTestStore(t)

	providers := ai.NewRegistry()
	reader := usercfg.NewAtomicReader(usercfg.Defaults())
	svc := newWalkthroughService(t, store, providers, reader)

	_, err := svc.Generate(context.Background(), GenerateRequest{
		State: repostate.RepositoryState{Root: "/r"},
		Kind:  "branch",
	})

	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("expected ErrProviderUnavailable, got %v", err)
	}
}

func TestWalkthroughGenerateReturnsCachedWhenFingerprintMatches(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	provider := &stubProvider{
		id: "anthropic",
		response: ai.GenerateResponse{
			ProviderID: "anthropic",
			ModelID:    "stub",
			Text:       `{"summary":"s","groups":[]}`,
		},
	}

	providers := ai.NewRegistry()
	providers.Register(provider)
	reader := usercfg.NewAtomicReader(usercfg.Defaults())
	svc := newWalkthroughService(t, store, providers, reader)

	// Pre-seed a walkthrough whose Fingerprint matches the one the service
	// will compute. Because the service derives the fingerprint from
	// (state, provider.ID()), we can construct it the same way.
	state := repostate.RepositoryState{Root: "/r"}
	expectedFP := walksFingerprint(state, provider.ID())

	if err := store.Walkthroughs.UpsertWalkthrough(ctx, storage.WalkthroughRecord{
		RepoRoot:    "/r",
		ContextKind: "branch",
		ContextSHA:  "sha",
		Fingerprint: expectedFP,
		ProviderID:  "anthropic",
		ModelID:     "stub",
		Summary:     "cached",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	result, err := svc.Generate(ctx, GenerateRequest{
		State:      state,
		Kind:       "branch",
		ContextSHA: "sha",
	})

	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if !result.Cached {
		t.Fatalf("expected cache hit, got %+v", result)
	}

	if result.Record.Summary != "cached" {
		t.Fatalf("expected cached summary, got %q", result.Record.Summary)
	}
}

func TestWalkthroughGenerateBypassesCacheOnRefresh(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	provider := &stubProvider{
		id: "anthropic",
		response: ai.GenerateResponse{
			ProviderID: "anthropic",
			ModelID:    "stub",
			Text:       `{"summary":"fresh","groups":[]}`,
		},
	}

	providers := ai.NewRegistry()
	providers.Register(provider)
	reader := usercfg.NewAtomicReader(usercfg.Defaults())
	svc := newWalkthroughService(t, store, providers, reader)

	state := repostate.RepositoryState{Root: "/r"}
	expectedFP := walksFingerprint(state, provider.ID())

	if err := store.Walkthroughs.UpsertWalkthrough(ctx, storage.WalkthroughRecord{
		RepoRoot:    "/r",
		ContextKind: "branch",
		ContextSHA:  "sha",
		Fingerprint: expectedFP,
		ProviderID:  "anthropic",
		ModelID:     "stub",
		Summary:     "cached",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	result, err := svc.Generate(ctx, GenerateRequest{
		State:      state,
		Kind:       "branch",
		ContextSHA: "sha",
		Refresh:    true,
	})

	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if result.Cached {
		t.Fatalf("expected cache bypass on refresh, got Cached=true")
	}
}
