package ai

import "strings"

// Internal helpers shared across provider implementations. Kept here so
// each provider stays focused on its wire protocol, not the cross-cutting
// "did the caller supply a model?" / "what prompt do we actually send?"
// logic.

// userOrDefaultModel returns the caller's preference when supplied, the
// provider default otherwise.
func (req GenerateRequest) userOrDefaultModel(fallback string) string {
	if m := strings.TrimSpace(req.Model); m != "" {
		return m
	}

	return fallback
}

// cappedMaxTokens normalises the MaxTokens field: zero or negative means
// "use the provider default supplied by the caller".
func (req GenerateRequest) cappedMaxTokens(defaultMax int) int {
	if req.MaxTokens > 0 {
		return req.MaxTokens
	}

	return defaultMax
}

// combinedPrompt returns the user prompt with the system prompt prepended
// when present. Providers that natively support a system role override
// this in their own Generate.
func (req GenerateRequest) combinedPrompt() string {
	system := strings.TrimSpace(req.SystemPrompt)
	user := strings.TrimSpace(req.UserPrompt)

	if system == "" {
		return user
	}

	if user == "" {
		return system
	}

	return system + "\n\n" + user
}
