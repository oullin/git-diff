package ai

import "strings"

func (req GenerateRequest) userOrDefaultModel(fallback string) string {
	if m := strings.TrimSpace(req.Model); m != "" {
		return m
	}

	return fallback
}

// cappedMaxTokens treats zero or negative as "use the caller's default".
func (req GenerateRequest) cappedMaxTokens(defaultMax int) int {
	if req.MaxTokens > 0 {
		return req.MaxTokens
	}

	return defaultMax
}

// combinedPrompt is the fallback for providers with no native system
// role; those that have one should build the request themselves.
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
