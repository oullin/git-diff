package walkthrough

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocanto/git-diff/internal/ai"
	"github.com/gocanto/git-diff/internal/review"
)

// Request configures one Generate call. Provider is required — callers
// look it up from an ai.Registry using the user-config-selected ID.
// Budget is optional; an empty Budget gets DefaultBudget().
type Request struct {
	Provider ai.Provider
	State    review.RepositoryState
	Budget   Budget
}

// ErrProviderRequired indicates the caller forgot to wire a provider.
// Distinct sentinel so the service layer can surface a 412 instead of a
// generic 500.
var ErrProviderRequired = errors.New("walkthrough requires an ai.Provider")

// Generate is the thin orchestrator: budget → prompt → provider → parse
// → validate → assemble. Each step lives in its own file so this
// function reads top-to-bottom in one screen.
//
// Empty file lists short-circuit (no model call) — saves a token spend
// for trivially clean states.
func Generate(ctx context.Context, req Request) (Walkthrough, error) {
	if req.Provider == nil {
		return Walkthrough{}, ErrProviderRequired
	}

	budget := req.Budget

	if budget.PerFileBytes == 0 && budget.TotalBytes == 0 {
		budget = DefaultBudget()
	}

	if len(req.State.Files) == 0 {
		return emptyWalkthrough(req.Provider), nil
	}

	prompt := BuildPrompt(PromptInput{State: req.State, Budget: budget})

	resp, err := req.Provider.Generate(ctx, ai.GenerateRequest{
		UserPrompt: prompt,
		MaxTokens:  defaultMaxTokens,
	})

	if err != nil {
		return Walkthrough{}, fmt.Errorf("provider %s: %w", req.Provider.ID(), err)
	}

	parsed, err := Parse(resp.Text)

	if err != nil {
		return Walkthrough{}, err
	}

	validated, err := Validate(parsed, req.State)

	if err != nil {
		return Walkthrough{}, err
	}

	order, notes := flattenForLegacy(validated.Groups)

	return Walkthrough{
		Groups:      validated.Groups,
		Summary:     validated.Summary,
		ProviderID:  resp.ProviderID,
		ModelID:     resp.ModelID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Fingerprint: FingerprintForStateAndProvider(req.State, resp.ProviderID),
		Order:       order,
		Notes:       notes,
	}, nil
}

const defaultMaxTokens = 2048

func emptyWalkthrough(provider ai.Provider) Walkthrough {
	return Walkthrough{
		Groups:      []Group{},
		Order:       []string{},
		Notes:       map[string]string{},
		ProviderID:  provider.ID(),
		ModelID:     provider.DefaultModel(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}
