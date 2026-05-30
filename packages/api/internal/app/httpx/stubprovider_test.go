package httpx

import (
	"context"
	"errors"

	"github.com/oullin/git-diff/internal/ai"
)

// stubProvider is a deterministic ai.Provider used in tests. It returns the
// Text/Err configured on construction or via WithResponse.
type stubProvider struct {
	id       string
	model    string
	response ai.GenerateResponse
	err      error
}

func newStubProvider(id string) *stubProvider {
	return &stubProvider{
		id:    id,
		model: "test-model",
		response: ai.GenerateResponse{
			ProviderID: id,
			ModelID:    "test-model",
			Text:       "{}",
		},
	}
}

func (p *stubProvider) ID() string { return p.id }

func (p *stubProvider) DefaultModel() string { return p.model }

func (p *stubProvider) SupportsModel(string) bool { return true }

func (p *stubProvider) Generate(ctx context.Context, _ ai.GenerateRequest) (ai.GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return ai.GenerateResponse{}, err
	}

	if p.err != nil {
		return ai.GenerateResponse{}, p.err
	}

	return p.response, nil
}

// withResponse swaps the canned response.
func (p *stubProvider) withResponse(text string) *stubProvider {
	p.response.Text = text

	return p
}

// withError makes Generate fail with err.
func (p *stubProvider) withError(err error) *stubProvider {
	if err == nil {
		err = errors.New("stub provider error")
	}

	p.err = err

	return p
}
