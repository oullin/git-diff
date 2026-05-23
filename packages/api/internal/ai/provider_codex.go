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

	"github.com/gocanto/git-diff/internal/usercfg"
)

// CodexProvider shells out to the `codex` CLI. When the binary is
// missing, Generate returns ErrCodexNotInstalled with no silent fallback.
type CodexProvider struct {
	cfg usercfg.Reader

	// resolveBinary lets tests inject a fake lookup.
	resolveBinary func() (string, error)

	// runExec lets tests inject a fake subprocess.
	runExec func(ctx context.Context, binary, model, prompt string) ([]byte, error)
}

// ErrCodexNotInstalled is returned when the `codex` binary can't be
// located. The HTTP layer maps this to a 412 so the renderer can prompt
// the user to install Codex or switch providers in their YAML config.
var ErrCodexNotInstalled = errors.New("codex CLI not found; install from https://github.com/openai/codex or switch walkthrough.provider in ~/.git-diff/config.yaml")

func NewCodexProvider(cfg usercfg.Reader) *CodexProvider {
	return &CodexProvider{
		cfg:           cfg,
		resolveBinary: defaultResolveCodexBinary,
		runExec:       defaultRunCodex,
	}
}

func (*CodexProvider) ID() string { return "codex" }

func (p *CodexProvider) DefaultModel() string { return p.cfg.Get().Codex.DefaultModel }

// SupportsModel is permissive: gatekeeping a hard-coded model list here
// would just race new OpenAI releases.
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

// defaultRunCodex propagates stderr in the error message so users see
// install/auth issues directly.
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
