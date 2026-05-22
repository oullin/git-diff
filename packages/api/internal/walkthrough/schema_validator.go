package walkthrough

import (
	"fmt"

	"github.com/gocanto/git-diff/internal/review"
)

// Validate trims a parsedResponse so it only references paths present in
// state, normalises invalid action/impact values to sensible defaults,
// and returns an error when the response is unusable (empty groups,
// every path filtered).
//
// Single responsibility: filter + normalise. Parsing happens in
// response_parser.go; persistence in cache.go.
func Validate(parsed parsedResponse, state review.RepositoryState) (parsedResponse, error) {
	allowed := allowedPaths(state)

	cleanGroups := parsed.Groups[:0]

	for _, group := range parsed.Groups {
		cleanFiles := group.Files[:0]

		for _, file := range group.Files {
			if _, ok := allowed[file.Path]; !ok {
				continue // drop hallucinated path
			}

			file.Action = normaliseAction(file.Action)
			file.Impact = normaliseImpact(file.Impact)
			cleanFiles = append(cleanFiles, file)
		}

		if len(cleanFiles) == 0 {
			continue // drop now-empty group
		}

		group.Files = cleanFiles
		cleanGroups = append(cleanGroups, group)
	}

	parsed.Groups = cleanGroups

	if len(parsed.Groups) == 0 {
		return parsedResponse{}, fmt.Errorf("walkthrough response had no valid file entries")
	}

	return parsed, nil
}

func allowedPaths(state review.RepositoryState) map[string]struct{} {
	allowed := make(map[string]struct{}, len(state.Files))

	for _, file := range state.Files {
		allowed[file.Path] = struct{}{}
	}

	return allowed
}

func normaliseAction(a Action) Action {
	switch a {
	case ActionReview, ActionScan, ActionSkim:
		return a
	default:
		return ActionReview
	}
}

func normaliseImpact(i Impact) Impact {
	switch i {
	case ImpactWide, ImpactContained, ImpactMechanical:
		return i
	default:
		return ImpactContained
	}
}

// flattenForLegacy mirrors Groups back onto the flat Order/Notes shape
// so legacy renderer code keeps working for one release. Pure function.
func flattenForLegacy(groups []Group) (order []string, notes map[string]string) {
	order = []string{}
	notes = map[string]string{}

	for _, group := range groups {
		for _, file := range group.Files {
			order = append(order, file.Path)
			notes[file.Path] = file.Note
		}
	}

	return order, notes
}
