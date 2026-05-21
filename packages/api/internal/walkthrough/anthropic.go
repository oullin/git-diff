package walkthrough

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MessagesRequest captures the inputs the Anthropic Messages API needs for
// a one-shot walkthrough call. Headers (api version, content type) are the
// implementation's concern.
type MessagesRequest struct {
	APIKey    string
	ModelID   string
	MaxTokens int
	Prompt    string
}

// MessagesResponse is the relevant slice of the Anthropic response that
// Generate consumes — the text of the first non-empty `text` block.
type MessagesResponse struct {
	Text string
}

// AnthropicClient calls the Anthropic Messages API. Tests swap the default
// implementation via SetClient to avoid real network calls.
type AnthropicClient interface {
	Messages(ctx context.Context, req MessagesRequest) (MessagesResponse, error)
}

type httpAnthropicClient struct {
	http *http.Client
}

func newHTTPAnthropicClient() *httpAnthropicClient {
	return &httpAnthropicClient{http: &http.Client{Timeout: 60 * time.Second}}
}

func (c *httpAnthropicClient) Messages(ctx context.Context, req MessagesRequest) (MessagesResponse, error) {
	payload := map[string]any{
		"model":      req.ModelID,
		"max_tokens": req.MaxTokens,
		"messages": []map[string]any{
			{"role": "user", "content": req.Prompt},
		},
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return MessagesResponse{}, fmt.Errorf("encode walkthrough request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, bytes.NewReader(body))

	if err != nil {
		return MessagesResponse{}, fmt.Errorf("build walkthrough request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", apiVersion)
	httpReq.Header.Set("x-api-key", req.APIKey)

	resp, err := c.http.Do(httpReq)

	if err != nil {
		return MessagesResponse{}, fmt.Errorf("call anthropic messages: %w", err)
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		return MessagesResponse{}, fmt.Errorf("read walkthrough response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return MessagesResponse{}, fmt.Errorf("anthropic messages returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	text, err := extractText(raw)

	if err != nil {
		return MessagesResponse{}, err
	}

	return MessagesResponse{Text: text}, nil
}

// Package-level Anthropic client. Tests use SetClient for scoped overrides.
var client AnthropicClient = newHTTPAnthropicClient()

// SetClient swaps the active Anthropic client and returns a restore func.
func SetClient(c AnthropicClient) (restore func()) {
	prev := client
	client = c

	return func() { client = prev }
}

func extractText(body []byte) (string, error) {
	var resp anthropicResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return block.Text, nil
		}
	}

	return "", errors.New("anthropic response had no text content")
}
