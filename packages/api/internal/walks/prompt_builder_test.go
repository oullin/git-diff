package walks

import (
	"strings"
	"testing"

	"github.com/oullin/git-diff/internal/review"
)

func TestBuildPromptIncludesInstructionsAndContext(t *testing.T) {
	state := review.RepositoryState{
		Root:      "/r",
		Mode:      "working",
		CommitSHA: "abc123",
		Files: []review.ChangedFile{
			{Path: "a.go", Status: review.StatusModified, Additions: 2, Deletions: 1,
				Sections: []review.DiffSection{{Patch: "+ hi\n- bye"}}},
		},
	}

	out := BuildPrompt(PromptInput{State: state, Budget: DefaultBudget()})

	if !strings.Contains(out, "Repository: /r") {
		t.Fatalf("expected repo line, got %q", out)
	}

	if !strings.Contains(out, "Mode: working") {
		t.Fatalf("expected mode line, got %q", out)
	}

	if !strings.Contains(out, "Commit: abc123") {
		t.Fatalf("expected commit line, got %q", out)
	}

	if !strings.Contains(out, "a.go") {
		t.Fatalf("expected file path in prompt, got %q", out)
	}

	if !strings.Contains(out, "JSON object") {
		t.Fatalf("expected instructions block, got %q", out)
	}
}

func TestBuildPromptOmitsCommitWhenEmpty(t *testing.T) {
	state := review.RepositoryState{
		Root: "/r", Mode: "working",
		Files: []review.ChangedFile{{Path: "a.go", Sections: []review.DiffSection{{Patch: "+hi"}}}},
	}

	out := BuildPrompt(PromptInput{State: state, Budget: DefaultBudget()})

	if strings.Contains(out, "Commit:") {
		t.Fatalf("expected no Commit line when CommitSHA empty, got %q", out)
	}
}

func TestBuildPromptAddsOmittedMarkerWhenBudgetCutsTail(t *testing.T) {
	state := review.RepositoryState{
		Files: []review.ChangedFile{
			{Path: "a.go", Sections: []review.DiffSection{{Patch: strings.Repeat("a", 500)}}},
			{Path: "b.go", Sections: []review.DiffSection{{Patch: strings.Repeat("b", 500)}}},
			{Path: "c.go", Sections: []review.DiffSection{{Patch: strings.Repeat("c", 500)}}},
		},
	}

	out := BuildPrompt(PromptInput{State: state, Budget: Budget{PerFileBytes: 4096, TotalBytes: 600}})

	if !strings.Contains(out, "remaining files omitted") {
		t.Fatalf("expected omitted marker for over-budget tail, got %q", out)
	}
}
