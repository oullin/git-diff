package walks

import (
	"strings"
	"testing"

	"github.com/oullin/git-diff/internal/domain/repostate"
)

func TestBuildPromptIncludesInstructionsAndContext(t *testing.T) {
	state := repostate.RepositoryState{
		Root:      "/r",
		Mode:      "working",
		CommitSHA: "abc123",
		Files: []repostate.ChangedFile{
			{Path: "a.go", Status: repostate.StatusModified, Additions: 2, Deletions: 1,
				Sections: []repostate.DiffSection{{Patch: "+ hi\n- bye"}}},
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
	state := repostate.RepositoryState{
		Root: "/r", Mode: "working",
		Files: []repostate.ChangedFile{{Path: "a.go", Sections: []repostate.DiffSection{{Patch: "+hi"}}}},
	}

	out := BuildPrompt(PromptInput{State: state, Budget: DefaultBudget()})

	if strings.Contains(out, "Commit:") {
		t.Fatalf("expected no Commit line when CommitSHA empty, got %q", out)
	}
}

func TestBuildPromptAddsOmittedMarkerWhenBudgetCutsTail(t *testing.T) {
	state := repostate.RepositoryState{
		Files: []repostate.ChangedFile{
			{Path: "a.go", Sections: []repostate.DiffSection{{Patch: strings.Repeat("a", 500)}}},
			{Path: "b.go", Sections: []repostate.DiffSection{{Patch: strings.Repeat("b", 500)}}},
			{Path: "c.go", Sections: []repostate.DiffSection{{Patch: strings.Repeat("c", 500)}}},
		},
	}

	out := BuildPrompt(PromptInput{State: state, Budget: Budget{PerFileBytes: 4096, TotalBytes: 600}})

	if !strings.Contains(out, "remaining files omitted") {
		t.Fatalf("expected omitted marker for over-budget tail, got %q", out)
	}
}
