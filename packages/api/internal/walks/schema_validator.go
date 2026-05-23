package walks

import (
	"fmt"

	"github.com/gocanto/git-diff/internal/review"
)

// Validate keeps only paths present in state, normalises action/impact
// to known values, and errors when nothing usable remains.
func Validate(parsed parsedResponse, state review.RepositoryState) (parsedResponse, error) {
	allowed := allowedPaths(state)

	cleanGroups := parsed.Groups[:0]

	for _, group := range parsed.Groups {
		cleanFiles := group.Files[:0]

		for _, file := range group.Files {
			if _, ok := allowed[file.Path]; !ok {
				continue
			}

			file.Action = normaliseAction(file.Action)
			file.Impact = normaliseImpact(file.Impact)
			cleanFiles = append(cleanFiles, file)
		}

		if len(cleanFiles) == 0 {
			continue
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
