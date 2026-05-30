package walks

import (
	"testing"

	"github.com/oullin/git-diff/internal/review"
)

func TestValidateDropsUnknownPaths(t *testing.T) {
	state := review.RepositoryState{Files: []review.ChangedFile{{Path: "a.go"}}}

	parsed := parsedResponse{
		Summary: "x",
		Groups: []Group{
			{
				ID:    "g",
				Title: "T",
				Files: []FileEntry{
					{Path: "a.go", Action: ActionReview, Impact: ImpactWide},
					{Path: "ghost.go", Action: ActionReview, Impact: ImpactWide},
				},
			},
		},
	}

	out, err := Validate(parsed, state)

	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if len(out.Groups) != 1 || len(out.Groups[0].Files) != 1 {
		t.Fatalf("expected single file kept, got %#v", out.Groups)
	}

	if out.Groups[0].Files[0].Path != "a.go" {
		t.Fatalf("expected a.go kept, got %q", out.Groups[0].Files[0].Path)
	}
}

func TestValidateRemovesEmptyGroups(t *testing.T) {
	state := review.RepositoryState{Files: []review.ChangedFile{{Path: "a.go"}}}

	parsed := parsedResponse{
		Groups: []Group{
			{ID: "keep", Files: []FileEntry{{Path: "a.go", Action: ActionReview, Impact: ImpactWide}}},
			{ID: "drop", Files: []FileEntry{{Path: "missing.go"}}},
		},
	}

	out, err := Validate(parsed, state)

	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if len(out.Groups) != 1 || out.Groups[0].ID != "keep" {
		t.Fatalf("expected only 'keep' group, got %#v", out.Groups)
	}
}

func TestValidateErrorsWhenNothingRemains(t *testing.T) {
	state := review.RepositoryState{Files: []review.ChangedFile{{Path: "a.go"}}}

	parsed := parsedResponse{
		Groups: []Group{
			{ID: "g", Files: []FileEntry{{Path: "missing.go"}}},
		},
	}

	if _, err := Validate(parsed, state); err == nil {
		t.Fatalf("expected error when no entries remain")
	}
}

func TestValidateNormalisesActionAndImpact(t *testing.T) {
	state := review.RepositoryState{Files: []review.ChangedFile{{Path: "a.go"}}}

	parsed := parsedResponse{
		Groups: []Group{
			{ID: "g", Files: []FileEntry{{Path: "a.go", Action: "nonsense", Impact: "weird"}}},
		},
	}

	out, _ := Validate(parsed, state)

	if out.Groups[0].Files[0].Action != ActionReview {
		t.Fatalf("expected unknown action -> review, got %q", out.Groups[0].Files[0].Action)
	}

	if out.Groups[0].Files[0].Impact != ImpactContained {
		t.Fatalf("expected unknown impact -> contained, got %q", out.Groups[0].Files[0].Impact)
	}
}
