// Package ai owns the AI-provider abstraction the rest of the app
// depends on for LLM calls. Single Provider interface, plug-in
// implementations (Anthropic, Codex CLI, future OpenAI/Bedrock/Ollama).
// Callers depend on the interface, not the concrete — Open/Closed +
// Dependency Inversion.
package ai

import "context"

// Usage is the token accounting a provider returns when available. All
// fields are best-effort — providers that can't report usage zero them.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GenerateRequest is the cross-provider payload. Providers are free to
// ignore fields that don't apply (e.g. JSONSchema for one-shot completion
// providers that can't constrain output format).
type GenerateRequest struct {
	// SystemPrompt is the system-role message. Some providers concatenate
	// with UserPrompt internally; that's their concern.
	SystemPrompt string

	// UserPrompt is the user-role message. Required.
	UserPrompt string

	// Model is the caller's preferred model. Providers that recognise the
	// string use it; an empty Model falls back to Provider.DefaultModel.
	Model string

	// MaxTokens caps the response length. Providers default sensibly when
	// zero.
	MaxTokens int

	// JSONSchema is an optional response schema. Providers that support
	// structured output use it; others ignore it.
	JSONSchema any
}

// GenerateResponse is the cross-provider result. Providers fill the
// optional fields they can; consumers must tolerate empty Usage.
type GenerateResponse struct {
	// ProviderID echoes Provider.ID() — handy for logs/metrics.
	ProviderID string

	// ModelID is the resolved model name actually used (post-validation).
	ModelID string

	// Text is the model's response, raw. Parsing/validation is the
	// caller's responsibility — providers don't interpret content.
	Text string

	Usage Usage
}

// Provider is the narrow interface every backend implements. Keep it
// small — when a feature needs streaming, embeddings, or tool use,
// split into a sibling interface (Streamer, Embedder, Tooler) rather
// than widening this one.
type Provider interface {
	// ID is the stable identifier used by config + registry lookup
	// ("anthropic", "codex"). Must be lowercase, no whitespace.
	ID() string

	// DefaultModel is the model the provider falls back to when the user
	// config doesn't specify one (or specifies the empty string).
	DefaultModel() string

	// SupportsModel returns true when the provider can serve the model
	// string. Providers should be permissive (accept new model names
	// they don't recognise) — the goal is to catch wrong-provider
	// mistakes, not to gatekeep new releases.
	SupportsModel(model string) bool

	// Generate runs a single one-shot completion. Honours context
	// cancellation. Returns the response text in GenerateResponse.Text;
	// callers parse/validate.
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}
