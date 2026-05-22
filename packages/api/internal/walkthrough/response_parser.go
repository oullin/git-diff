package walkthrough

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parsedResponse is the JSON shape we expect back from the model.
// Mirrors PromptInput's documented schema 1:1.
type parsedResponse struct {
	Summary string  `json:"summary"`
	Groups  []Group `json:"groups"`
	// Legacy fields — populated when the model (or a cached row) still
	// emits the flat shape. Converted to groups by Parse.
	Order []string          `json:"order"`
	Notes map[string]string `json:"notes"`
}

// Parse decodes the model's output into a Walkthrough's payload fields
// (Groups, Summary). Strips optional ```json fences first.
//
// No validation happens here — the SchemaValidator runs separately so
// each unit stays focused. Parse will accept anything the JSON decoder
// accepts; validation is the gate.
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

	// Fall through to legacy: if the model emitted Order/Notes instead
	// of Groups, hoist the flat list into one group so downstream code
	// sees the same shape.
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
