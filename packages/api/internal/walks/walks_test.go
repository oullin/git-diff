package walks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/review"
)

type mockProvider struct {
	text              string
	called            bool
	shouldNotBeCalled bool
}

func TestFingerprintForStateIsStableAndOrderSensitive(t *testing.T) {
	build := func() review.RepositoryState {
		return review.RepositoryState{
			Mode:    review.RepositoryModeWorking,
			HeadSHA: "abc",
			Files: []review.ChangedFile{
				{Path: "a.go", Fingerprint: "fp-a"},
				{Path: "b.go", Fingerprint: "fp-b"},
			},
		}
	}

	a := build()
	b := build()

	if FingerprintForState(a) != FingerprintForState(b) {
		t.Fatal("identical states should produce equal fingerprints")
	}

	a.Files[0].Fingerprint = "fp-a-2"

	if FingerprintForState(a) == FingerprintForState(b) {
		t.Fatal("changing a fingerprint must change the state fingerprint")
	}
}

func TestFingerprintMixesProviderID(t *testing.T) {
	state := review.RepositoryState{
		Mode:  review.RepositoryModeWorking,
		Files: []review.ChangedFile{{Path: "x.go", Fingerprint: "fp"}},
	}

	if FingerprintForStateAndProvider(state, "anthropic") == FingerprintForStateAndProvider(state, "codex") {
		t.Fatal("switching providers must invalidate the fingerprint")
	}
}

func TestBuildPromptIncludesPathsAndTruncates(t *testing.T) {
	state := review.RepositoryState{
		Root: "/repo",
		Mode: review.RepositoryModeWorking,
		Files: []review.ChangedFile{
			{
				Path: "src/main.ts",
				Sections: []review.DiffSection{
					{Kind: "unstaged", Patch: "@@ -1 +1,2 @@\n+console.log('hi')\n"},
				},
			},
		},
	}

	prompt := BuildPrompt(PromptInput{State: state, Budget: Budget{PerFileBytes: 10, TotalBytes: 1000}})

	if !strings.Contains(prompt, "src/main.ts") {
		t.Fatal("expected path to appear in prompt")
	}

	if !strings.Contains(prompt, "truncated") {
		t.Fatalf("per-file budget should truncate; prompt =\n%s", prompt)
	}
}

func TestEnforceBudgetStopsAtTotalCap(t *testing.T) {
	files := []review.ChangedFile{
		{Path: "a", Sections: []review.DiffSection{{Patch: strings.Repeat("x", 40)}}},
		{Path: "b", Sections: []review.DiffSection{{Patch: strings.Repeat("y", 40)}}},
		{Path: "c", Sections: []review.DiffSection{{Patch: strings.Repeat("z", 40)}}},
	}

	got := enforceBudget(files, Budget{PerFileBytes: 100, TotalBytes: 50})

	if len(got) < 1 {
		t.Fatalf("expected at least one file before the cap, got %d", len(got))
	}

	if !got[len(got)-1].OmittedAfter {
		t.Fatal("last included file should flag OmittedAfter when total cap hits")
	}
}

func TestParseDecodesGroupsShape(t *testing.T) {
	newShape := `{"summary":"refactor","groups":[{"id":"g1","title":"Auth","rationale":"core","files":[{"path":"a.go","note":"check","action":"review","impact":"contained"}]}]}`

	got, err := Parse(newShape)

	if err != nil {
		t.Fatal(err)
	}

	if len(got.Groups) != 1 || got.Groups[0].Files[0].Path != "a.go" {
		t.Fatalf("groups-shape parse failed: %+v", got)
	}
}

func TestParseLeavesGroupsEmptyForLegacyOrderNotes(t *testing.T) {
	legacy := "```json\n" + `{"summary":"flat","order":["x.go","y.go"],"notes":{"x.go":"first"}}` + "\n```"

	got, err := Parse(legacy)

	if err != nil {
		t.Fatal(err)
	}

	if len(got.Groups) != 0 {
		t.Fatalf("legacy order/notes should no longer produce groups: %+v", got)
	}
}

func TestValidateDropsHallucinatedPathsAndNormalisesEnums(t *testing.T) {
	state := review.RepositoryState{
		Files: []review.ChangedFile{{Path: "real.go"}},
	}

	parsed := parsedResponse{
		Groups: []Group{
			{
				ID:    "g1",
				Title: "Mixed",
				Files: []FileEntry{
					{Path: "real.go", Action: "weird", Impact: "huge"},
					{Path: "fake.go", Action: ActionScan, Impact: ImpactWide},
				},
			},
		},
	}

	got, err := Validate(parsed, state)

	if err != nil {
		t.Fatal(err)
	}

	if len(got.Groups) != 1 || len(got.Groups[0].Files) != 1 {
		t.Fatalf("expected fake.go dropped, real.go kept: %+v", got)
	}

	file := got.Groups[0].Files[0]

	if file.Action != ActionReview || file.Impact != ImpactContained {
		t.Fatalf("expected enum fallbacks; got action=%q impact=%q", file.Action, file.Impact)
	}
}

func TestValidateRejectsAllFiltered(t *testing.T) {
	state := review.RepositoryState{
		Files: []review.ChangedFile{{Path: "real.go"}},
	}

	parsed := parsedResponse{
		Groups: []Group{
			{Files: []FileEntry{{Path: "fake1.go"}, {Path: "fake2.go"}}},
		},
	}

	_, err := Validate(parsed, state)

	if err == nil {
		t.Fatal("expected error when every path is filtered")
	}
}

func TestGenerateOrchestratesProviderParseValidate(t *testing.T) {
	provider := &mockProvider{
		text: `{"summary":"theme","groups":[{"id":"g","title":"T","rationale":"r","files":[{"path":"x.go","note":"check","action":"review","impact":"contained"}]}]}`,
	}

	state := review.RepositoryState{
		Mode:  review.RepositoryModeWorking,
		Files: []review.ChangedFile{{Path: "x.go"}},
	}

	got, err := Generate(context.Background(), Request{Provider: provider, State: state})

	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if len(got.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got.Groups))
	}

	if got.ProviderID != "mock" {
		t.Fatalf("provider id = %q", got.ProviderID)
	}
}

func TestGenerateRequiresProvider(t *testing.T) {
	_, err := Generate(context.Background(), Request{State: review.RepositoryState{}})

	if !errors.Is(err, ErrProviderRequired) {
		t.Fatalf("err = %v, want ErrProviderRequired", err)
	}
}

func TestGenerateShortCircuitsOnEmptyState(t *testing.T) {
	provider := &mockProvider{shouldNotBeCalled: true}

	got, err := Generate(context.Background(), Request{Provider: provider, State: review.RepositoryState{}})

	if err != nil {
		t.Fatal(err)
	}

	if len(got.Groups) != 0 {
		t.Fatalf("expected no groups for empty state, got %d", len(got.Groups))
	}

	if provider.called {
		t.Fatal("provider should not be called for an empty state")
	}
}

func (m *mockProvider) ID() string                { return "mock" }
func (m *mockProvider) DefaultModel() string      { return "mock-1" }
func (m *mockProvider) SupportsModel(string) bool { return true }
func (m *mockProvider) Generate(_ context.Context, _ ai.GenerateRequest) (ai.GenerateResponse, error) {
	if m.shouldNotBeCalled {
		panic("provider should not have been called")
	}

	m.called = true

	return ai.GenerateResponse{ProviderID: m.ID(), ModelID: m.DefaultModel(), Text: m.text}, nil
}
