package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/oullin/git-diff/internal/usercfg"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func TestAnthropicSupportsModel(t *testing.T) {
	p := NewAnthropicProvider(testReader())

	cases := map[string]bool{
		"claude-sonnet-4-5": true,
		"claude-opus-9":     true,
		"CLAUDE-3":          true,
		"gpt-5":             false,
		"":                  false,
		"   ":               false,
	}

	for model, want := range cases {
		if got := p.SupportsModel(model); got != want {
			t.Fatalf("SupportsModel(%q) = %v, want %v", model, got, want)
		}
	}
}

func TestAnthropicProviderIdentity(t *testing.T) {
	p := NewAnthropicProvider(testReader())

	if p.ID() != "anthropic" {
		t.Fatalf("id = %q", p.ID())
	}

	if p.DefaultModel() != usercfg.Defaults().Anthropic.DefaultModel {
		t.Fatalf("default model = %q", p.DefaultModel())
	}
}

func TestAnthropicNonOKResponseReadsBoundedErrorBody(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	errLimit := usercfg.Defaults().Anthropic.ErrorBodyLimit

	p := &AnthropicProvider{
		cfg: testReader(),
		http: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(errLimit)+1024))),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	_, err := p.Generate(context.Background(), GenerateRequest{UserPrompt: "hello"})

	if err == nil {
		t.Fatalf("expected non-OK error")
	}

	if !strings.Contains(err.Error(), "anthropic returned 502") {
		t.Fatalf("error = %q, want status", err.Error())
	}

	if int64(len(err.Error())) > errLimit+256 {
		t.Fatalf("error length = %d, want bounded body", len(err.Error()))
	}
}

func TestAnthropicSuccessResponseRejectsOversizedBody(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	successLimit := usercfg.Defaults().Anthropic.SuccessBodyLimit

	p := &AnthropicProvider{
		cfg: testReader(),
		http: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				body := `{"content":[{"type":"text","text":"` + strings.Repeat("x", int(successLimit)+1) + `"}]}`

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	_, err := p.Generate(context.Background(), GenerateRequest{UserPrompt: "hello"})

	if !errors.Is(err, errAnthropicResponseTooLarge) {
		t.Fatalf("error = %v, want errAnthropicResponseTooLarge", err)
	}
}

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
