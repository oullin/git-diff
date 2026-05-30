package walks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/review"
)

// Request configures one Generate call. An empty Budget gets DefaultBudget().
type Request struct {
	Provider ai.Provider
	State    review.RepositoryState
	Budget   Budget
}

var ErrProviderRequired = errors.New("walkthrough requires an ai.Provider")

// Generate short-circuits on empty file lists to avoid spending tokens on
// trivially clean states.
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

	return Walkthrough{
		Groups:      validated.Groups,
		Summary:     validated.Summary,
		ProviderID:  resp.ProviderID,
		ModelID:     resp.ModelID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Fingerprint: FingerprintForStateAndProvider(req.State, resp.ProviderID),
	}, nil
}

const defaultMaxTokens = 2048

func emptyWalkthrough(provider ai.Provider) Walkthrough {
	return Walkthrough{
		Groups:      []Group{},
		ProviderID:  provider.ID(),
		ModelID:     provider.DefaultModel(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}
