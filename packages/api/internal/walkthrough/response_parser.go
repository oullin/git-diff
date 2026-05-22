package walkthrough

import (
	"encoding/json"
	"fmt"
	"strings"
)

type parsedResponse struct {
	Summary string  `json:"summary"`
	Groups  []Group `json:"groups"`
	// Order/Notes are the legacy flat shape; Parse hoists them into Groups.
	Order []string          `json:"order"`
	Notes map[string]string `json:"notes"`
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

	if len(parsed.Groups) == 0 && len(parsed.Order) > 0 {
		parsed.Groups = []Group{legacyToGroup(parsed.Order, parsed.Notes)}
	}

	return parsed, nil
}

func legacyToGroup(order []string, notes map[string]string) Group {
	files := make([]FileEntry, 0, len(order))

	for _, path := range order {
		files = append(files, FileEntry{
			Path:   path,
			Note:   notes[path],
			Action: ActionReview,
			Impact: ImpactContained,
		})
	}

	return Group{
		ID:        "files",
		Title:     "Files",
		Rationale: "",
		Files:     files,
	}
}
