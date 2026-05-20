package walkthrough

import (
	"strings"
	"testing"

	"github.com/gocanto/git-diff/internal/review"
)

func TestFingerprintForStateIsStableAndOrderSensitive(t *testing.T) {
	make := func() review.RepositoryState {
		return review.RepositoryState{
			Mode:    review.RepositoryModeWorking,
			HeadSHA: "abc",
			Files: []review.ChangedFile{
				{Path: "a.go", Fingerprint: "fp-a"},
				{Path: "b.go", Fingerprint: "fp-b"},
			},
		}
	}

	a := make()
	b := make()

	if FingerprintForState(a) != FingerprintForState(b) {
		t.Fatal("identical states should produce equal fingerprints")
	}

	a.Files[0].Fingerprint = "fp-a-2"

	if FingerprintForState(a) == FingerprintForState(b) {
		t.Fatal("changing a fingerprint must change the state fingerprint")
	}
}

func TestParseModelOutputFiltersPathsNotInState(t *testing.T) {
	state := review.RepositoryState{
		Files: []review.ChangedFile{{Path: "src/App.vue"}, {Path: "src/main.ts"}},
	}

	parsed, err := parseModelOutput(
		"```json\n"+`{"order":["src/main.ts","src/App.vue","fake.go"],"notes":{"src/main.ts":"entry","fake.go":"hallucinated"},"summary":"local"}`+"\n```",
		state,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(parsed.Order) != 2 || parsed.Order[0] != "src/main.ts" {
		t.Fatalf("order = %#v", parsed.Order)
	}

	if _, hallucinated := parsed.Notes["fake.go"]; hallucinated {
		t.Fatalf("notes leaked hallucination: %#v", parsed.Notes)
	}

	if parsed.Summary != "local" {
		t.Fatalf("summary = %q", parsed.Summary)
	}
}

func TestGenerateRejectsMissingAPIKey(t *testing.T) {
	_, err := Generate(nil, Request{State: review.RepositoryState{}})

	if err != ErrMissingAPIKey {
		t.Fatalf("err = %v, want ErrMissingAPIKey", err)
	}
}

func TestBuildPromptIncludesPathsAndPatches(t *testing.T) {
	prompt := buildPrompt(review.RepositoryState{
		Root: "/repo",
		Mode: review.RepositoryModeWorking,
		Files: []review.ChangedFile{
			{
				Path:      "src/main.ts",
				Status:    review.StatusModified,
				Additions: 3,
				Deletions: 1,
				Sections: []review.DiffSection{
					{Kind: "unstaged", Patch: "@@ -1 +1,2 @@\n+console.log('hi')\n"},
				},
			},
		},
	})

	if !strings.Contains(prompt, "src/main.ts") {
		t.Fatal("expected path to appear in prompt")
	}

	if !strings.Contains(prompt, "console.log") {
		t.Fatal("expected patch body to appear in prompt")
	}
}
