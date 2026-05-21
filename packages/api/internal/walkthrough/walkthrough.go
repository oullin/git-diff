// Package walkthrough integrates with the Anthropic Messages API to produce
// an ordered review walkthrough — a recommended file order plus a one-line
// note per file — from a RepositoryState.
package walkthrough

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/review"
)

// per-file truncation; keeps the prompt cheap

// Walkthrough is the structured output the renderer consumes. Order lists
// every file path the LLM recommends reviewing, in order; Notes maps each
// path to a short rationale.
type Walkthrough struct {
	Order       []string          `json:"order"`
	Notes       map[string]string `json:"notes"`
	Summary     string            `json:"summary,omitempty"`
	ModelID     string            `json:"modelId"`
	GeneratedAt string            `json:"generatedAt"`
	Fingerprint string            `json:"fingerprint"`
}

// Request configures Generate. APIKey is required; an empty key surfaces a
// distinguishable ErrMissingAPIKey so the caller can prompt the user.
type Request struct {
	APIKey  string
	ModelID string
	State   review.RepositoryState
}

// ErrMissingAPIKey is returned when no Anthropic API key was provided. The
// HTTP layer translates this to a 412 Precondition Failed so the UI knows to
// surface the settings panel.

// FingerprintForState returns a deterministic hash of the file paths and
// per-file fingerprints in the state. Two RepositoryStates with the same
// fingerprint should yield the same walkthrough; this powers the cache key
// for the storage layer.

// Generate calls the Anthropic Messages API with a structured prompt built
// from the repository state and returns a parsed Walkthrough.

type parsedOutput struct {
	Order   []string          `json:"order"`
	Notes   map[string]string `json:"notes"`
	Summary string            `json:"summary"`
}

// The model occasionally wraps the JSON in a ```json fence. Strip that.

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

const (
	defaultModelID   = "claude-sonnet-4-5"
	defaultMaxTokens = 1024
	apiEndpoint      = "https://api.anthropic.com/v1/messages"
	apiVersion       = "2023-06-01"
	maxPatchChars    = 3000
	totalPatchChars  = 60000
)

var ErrMissingAPIKey = errors.New("anthropic api key is required")

func FingerprintForState(state review.RepositoryState) string {
	hash := sha1.New()
	hash.Write([]byte(state.Mode))
	hash.Write([]byte(state.CommitSHA))
	hash.Write([]byte(state.HeadSHA))

	for _, file := range state.Files {
		hash.Write([]byte(file.Path))
		hash.Write([]byte(file.Fingerprint))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func Generate(ctx context.Context, req Request) (Walkthrough, error) {
	if strings.TrimSpace(req.APIKey) == "" {
		return Walkthrough{}, ErrMissingAPIKey
	}

	if len(req.State.Files) == 0 {
		return Walkthrough{
			Order:       []string{},
			Notes:       map[string]string{},
			ModelID:     resolveModelID(req.ModelID),
			GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Fingerprint: FingerprintForState(req.State),
		}, nil
	}

	prompt := buildPrompt(req.State)
	modelID := resolveModelID(req.ModelID)

	payload := map[string]any{
		"model":      modelID,
		"max_tokens": defaultMaxTokens,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return Walkthrough{}, fmt.Errorf("encode walkthrough request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, bytes.NewReader(body))

	if err != nil {
		return Walkthrough{}, fmt.Errorf("build walkthrough request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", apiVersion)
	httpReq.Header.Set("x-api-key", req.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)

	if err != nil {
		return Walkthrough{}, fmt.Errorf("call anthropic messages: %w", err)
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		return Walkthrough{}, fmt.Errorf("read walkthrough response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Walkthrough{}, fmt.Errorf("anthropic messages returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	text, err := extractText(raw)

	if err != nil {
		return Walkthrough{}, err
	}

	parsed, err := parseModelOutput(text, req.State)

	if err != nil {
		return Walkthrough{}, err
	}

	return Walkthrough{
		Order:       parsed.Order,
		Notes:       parsed.Notes,
		Summary:     parsed.Summary,
		ModelID:     modelID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Fingerprint: FingerprintForState(req.State),
	}, nil
}

func resolveModelID(override string) string {
	if id := strings.TrimSpace(override); id != "" {
		return id
	}

	if id := strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL")); id != "" {
		return id
	}

	return defaultModelID
}

func buildPrompt(state review.RepositoryState) string {
	var builder strings.Builder

	builder.WriteString(`You are helping a developer review a Git diff. Output ONLY a JSON object with three keys:
- "order": an array of file paths from the diff below, sorted in the order they should be reviewed.
- "notes": an object mapping each file path to a one-line note explaining why it should be reviewed in that position.
- "summary": one short sentence summarizing the overall theme of the change.

Use only paths from the provided list. Do not include any path that isn't in the diff. Output nothing outside the JSON object.

Repository: `)
	builder.WriteString(state.Root)
	builder.WriteString("\nMode: ")
	builder.WriteString(state.Mode)

	if state.CommitSHA != "" {
		builder.WriteString("\nCommit: ")
		builder.WriteString(state.CommitSHA)
	}

	builder.WriteString("\n\nChanged files (truncated patches):\n\n")

	total := 0

	for _, file := range state.Files {
		fmt.Fprintf(&builder, "----- %s (status=%s, +%d/-%d)\n", file.Path, file.Status, file.Additions, file.Deletions)

		var patch strings.Builder

		for _, section := range file.Sections {
			if section.Binary {
				patch.WriteString("[binary]\n")

				continue
			}

			patch.WriteString(section.Patch)
			patch.WriteString("\n")
		}

		body := patch.String()

		if len(body) > maxPatchChars {
			body = body[:maxPatchChars] + "\n…(truncated)\n"
		}

		total += len(body)

		if total > totalPatchChars {
			body += "\n…(remaining files omitted to stay within token budget)\n"
		}

		builder.WriteString(body)
		builder.WriteString("\n")

		if total > totalPatchChars {
			break
		}
	}

	return builder.String()
}

func parseModelOutput(text string, state review.RepositoryState) (parsedOutput, error) {
	trimmed := strings.TrimSpace(text)

	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var parsed parsedOutput

	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return parsedOutput{}, fmt.Errorf("decode walkthrough JSON: %w", err)
	}

	allowed := make(map[string]struct{}, len(state.Files))

	for _, file := range state.Files {
		allowed[file.Path] = struct{}{}
	}

	cleanOrder := parsed.Order[:0:len(parsed.Order)]

	for _, path := range parsed.Order {
		if _, ok := allowed[path]; ok {
			cleanOrder = append(cleanOrder, path)
		}
	}

	parsed.Order = cleanOrder

	if parsed.Notes == nil {
		parsed.Notes = map[string]string{}
	}

	for path := range parsed.Notes {
		if _, ok := allowed[path]; !ok {
			delete(parsed.Notes, path)
		}
	}

	return parsed, nil
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
