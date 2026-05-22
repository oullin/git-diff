package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicProvider implements Provider against the Anthropic Messages API.
// Pulls the API key from $ANTHROPIC_API_KEY at request time so key
// rotation doesn't require a process restart.
type AnthropicProvider struct {
	http *http.Client
}

// NewAnthropicProvider returns a provider with a 60 s default HTTP
// timeout. Tests can swap http via the field directly when needed.

// SupportsModel is permissive — any non-empty string that looks like a
// claude model name passes. Goal: catch obvious wrong-provider
// configs (e.g. "gpt-5"), not gatekeep new claude releases.

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

const (
	anthropicEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicAPIVer   = "2023-06-01"
	anthropicDefault  = "claude-sonnet-4-5"
)

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{http: &http.Client{Timeout: 60 * time.Second}}
}

func (*AnthropicProvider) ID() string { return "anthropic" }

func (*AnthropicProvider) DefaultModel() string { return anthropicDefault }

func (*AnthropicProvider) SupportsModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))

	if m == "" {
		return false
	}

	return strings.HasPrefix(m, "claude")
}

func (p *AnthropicProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))

	if apiKey == "" {
		return GenerateResponse{}, errors.New("ANTHROPIC_API_KEY is required for the anthropic provider")
	}

	model := strings.TrimSpace(req.userOrDefaultModel(p.DefaultModel()))

	if !p.SupportsModel(model) {
		return GenerateResponse{}, fmt.Errorf("anthropic provider does not support model %q (expected a claude-* name)", model)
	}

	payload := map[string]any{
		"model":      model,
		"max_tokens": req.cappedMaxTokens(1024),
		"messages": []map[string]any{
			{"role": "user", "content": req.combinedPrompt()},
		},
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return GenerateResponse{}, fmt.Errorf("encode anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(body))

	if err != nil {
		return GenerateResponse{}, fmt.Errorf("build anthropic request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", anthropicAPIVer)
	httpReq.Header.Set("x-api-key", apiKey)

	resp, err := p.http.Do(httpReq)

	if err != nil {
		return GenerateResponse{}, fmt.Errorf("call anthropic messages: %w", err)
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		return GenerateResponse{}, fmt.Errorf("read anthropic response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return GenerateResponse{}, fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	text, usage, err := decodeAnthropic(raw)

	if err != nil {
		return GenerateResponse{}, err
	}

	return GenerateResponse{
		ProviderID: p.ID(),
		ModelID:    model,
		Text:       text,
		Usage:      usage,
	}, nil
}

func decodeAnthropic(body []byte) (string, Usage, error) {
	var resp anthropicResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return "", Usage{}, fmt.Errorf("decode anthropic response: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return block.Text, Usage{InputTokens: resp.Usage.InputTokens, OutputTokens: resp.Usage.OutputTokens}, nil
		}
	}

	return "", Usage{}, errors.New("anthropic response had no text content")
}
