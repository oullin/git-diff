package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CodexProvider shells out to the `codex` CLI (the OpenAI Codex tool
// codiff uses for its walkthrough). Strict failure mode per the user's
// decision: when the binary isn't on PATH (and not under ~/.codex/bin),
// Generate returns ErrCodexNotInstalled — no silent fallback.
type CodexProvider struct {
	// resolveBinary lets tests inject a fake lookup. nil → default
	// resolution (PATH then ~/.codex/bin/codex).
	resolveBinary func() (string, error)

	// runExec lets tests inject a fake subprocess. nil → real os/exec.
	runExec func(ctx context.Context, binary, model, prompt string) ([]byte, error)
}

const (
	codexDefaultModel = "gpt-5.3-codex-spark"
)

// ErrCodexNotInstalled is returned when the `codex` binary can't be
// located. The HTTP layer maps this to a 412 so the renderer can prompt
// the user to install Codex or switch providers in their YAML config.
var ErrCodexNotInstalled = errors.New("codex CLI not found; install from https://github.com/openai/codex or switch walkthrough.provider in ~/.git-diff/config.yaml")

// NewCodexProvider returns a provider with the default binary resolution
// (PATH then ~/.codex/bin/codex) and a real subprocess runner.
func NewCodexProvider() *CodexProvider {
	return &CodexProvider{
		resolveBinary: defaultResolveCodexBinary,
		runExec:       defaultRunCodex,
	}
}

func (*CodexProvider) ID() string { return "codex" }

func (*CodexProvider) DefaultModel() string { return codexDefaultModel }

// SupportsModel is permissive — any non-empty string passes. Codex
// supports a wide range of OpenAI models; gatekeeping here would just
// race new releases.
func (*CodexProvider) SupportsModel(model string) bool {
	return strings.TrimSpace(model) != ""
}

func (p *CodexProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	binary, err := p.resolveBinary()

	if err != nil {
		return GenerateResponse{}, err
	}

	model := strings.TrimSpace(req.userOrDefaultModel(p.DefaultModel()))

	if !p.SupportsModel(model) {
		return GenerateResponse{}, fmt.Errorf("codex provider requires a non-empty model name")
	}

	prompt := req.combinedPrompt()

	if prompt == "" {
		return GenerateResponse{}, errors.New("codex provider requires a non-empty prompt")
	}

	output, err := p.runExec(ctx, binary, model, prompt)

	if err != nil {
		return GenerateResponse{}, err
	}

	return GenerateResponse{
		ProviderID: p.ID(),
		ModelID:    model,
		Text:       strings.TrimSpace(string(output)),
	}, nil
}

// defaultResolveCodexBinary looks up `codex` on PATH, then falls back to
// the install location codex typically uses. Returns ErrCodexNotInstalled
// when neither is present.
func defaultResolveCodexBinary() (string, error) {
	if path, err := exec.LookPath("codex"); err == nil {
		return path, nil
	}

	home, err := os.UserHomeDir()

	if err == nil {
		candidate := filepath.Join(home, ".codex", "bin", "codex")

		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", ErrCodexNotInstalled
}

// defaultRunCodex invokes the binary with `codex exec --model <m>
// --stdin` and captures stdout. Stderr is propagated in the error
// message so users see install / auth issues directly.
func defaultRunCodex(ctx context.Context, binary, model, prompt string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, "exec", "--model", model, "--stdin")
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codex exec: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}
