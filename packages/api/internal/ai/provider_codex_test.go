package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCodexProviderMissingBinaryReturnsSentinel(t *testing.T) {
	p := &CodexProvider{
		resolveBinary: func() (string, error) { return "", ErrCodexNotInstalled },
	}

	_, err := p.Generate(context.Background(), GenerateRequest{UserPrompt: "hi"})

	if !errors.Is(err, ErrCodexNotInstalled) {
		t.Fatalf("expected ErrCodexNotInstalled, got %v", err)
	}
}

func TestCodexProviderShellsOutWithUserPrompt(t *testing.T) {
	var sawPrompt, sawModel string

	p := &CodexProvider{
		resolveBinary: func() (string, error) { return "/fake/codex", nil },
		runExec: func(_ context.Context, _, model, prompt string) ([]byte, error) {
			sawPrompt = prompt
			sawModel = model

			return []byte("hello world"), nil
		},
	}

	resp, err := p.Generate(context.Background(), GenerateRequest{
		UserPrompt: "review this",
		Model:      "gpt-5",
	})

	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if resp.Text != "hello world" {
		t.Fatalf("text = %q", resp.Text)
	}

	if resp.ProviderID != "codex" || resp.ModelID != "gpt-5" {
		t.Fatalf("provider/model = %q/%q", resp.ProviderID, resp.ModelID)
	}

	if !strings.Contains(sawPrompt, "review this") {
		t.Fatalf("prompt missing user content: %q", sawPrompt)
	}

	if sawModel != "gpt-5" {
		t.Fatalf("model passed to exec = %q", sawModel)
	}
}

func TestCodexProviderFallsBackToDefaultModel(t *testing.T) {
	var captured string

	p := &CodexProvider{
		resolveBinary: func() (string, error) { return "/fake/codex", nil },
		runExec: func(_ context.Context, _, model, _ string) ([]byte, error) {
			captured = model

			return []byte("ok"), nil
		},
	}

	_, err := p.Generate(context.Background(), GenerateRequest{UserPrompt: "x"})

	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if captured != codexDefaultModel {
		t.Fatalf("expected default %q, got %q", codexDefaultModel, captured)
	}
}

func TestCodexProviderRejectsEmptyPrompt(t *testing.T) {
	p := &CodexProvider{
		resolveBinary: func() (string, error) { return "/fake/codex", nil },
		runExec:       func(context.Context, string, string, string) ([]byte, error) { return nil, nil },
	}

	_, err := p.Generate(context.Background(), GenerateRequest{})

	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}
