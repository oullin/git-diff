package walks

import (
	"encoding/json"
	"fmt"
	"strings"
)

type parsedResponse struct {
	Summary string  `json:"summary"`
	Groups  []Group `json:"groups"`
}

// Parse strips optional ```json fences and decodes; validation runs in
// Validate so this stays a tolerant JSON pass.
func Parse(raw string) (parsedResponse, error) {
	trimmed := strings.TrimSpace(raw)

	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var parsed parsedResponse

	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return parsedResponse{}, fmt.Errorf("decode walkthrough JSON: %w", err)
	}

	return parsed, nil
}
