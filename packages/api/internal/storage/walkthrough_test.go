package storage

import (
	"context"
	"testing"
)

func TestUpsertWalkthroughRoundTripsGroups(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	record := WalkthroughRecord{
		RepoRoot:    "/r",
		ContextKind: "branch",
		ContextSHA:  "sha",
		Fingerprint: "fp",
		ProviderID:  "anthropic",
		ModelID:     "claude-sonnet-4-5",
		Summary:     "summary",
		GeneratedAt: "2030-01-01T00:00:00Z",
		Groups: []WalkthroughGroup{
			{
				ID:        "g1",
				Title:     "Group 1",
				Rationale: "why",
				Files: []WalkthroughGroupFile{
					{Path: "a.go", Note: "n", Action: "edit", Impact: "low"},
				},
			},
		},
	}

	if err := store.Walkthroughs.UpsertWalkthrough(ctx, record); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, found, err := store.Walkthroughs.GetWalkthrough(ctx, "/r", "branch", "sha")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if !found {
		t.Fatalf("expected found=true")
	}

	if got.Summary != record.Summary {
		t.Fatalf("summary mismatch: %q vs %q", got.Summary, record.Summary)
	}

	if len(got.Groups) != 1 || got.Groups[0].ID != "g1" || len(got.Groups[0].Files) != 1 {
		t.Fatalf("groups did not round-trip: %#v", got.Groups)
	}

	if got.Groups[0].Files[0].Path != "a.go" {
		t.Fatalf("file path mismatch: %q", got.Groups[0].Files[0].Path)
	}
}

func TestGetWalkthroughReturnsFoundFalseForMissing(t *testing.T) {
	store := newTestStore(t)

	_, found, err := store.Walkthroughs.GetWalkthrough(context.Background(), "/r", "branch", "missing")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if found {
		t.Fatalf("expected found=false")
	}
}

func TestUpsertWalkthroughOverwritesExisting(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	first := WalkthroughRecord{
		RepoRoot: "/r", ContextKind: "branch", ContextSHA: "sha",
		Summary: "v1", ProviderID: "anthropic", ModelID: "m",
		Groups: []WalkthroughGroup{{ID: "g", Title: "v1"}},
	}

	if err := store.Walkthroughs.UpsertWalkthrough(ctx, first); err != nil {
		t.Fatalf("first: %v", err)
	}

	second := first
	second.Summary = "v2"
	second.Groups = []WalkthroughGroup{{ID: "g", Title: "v2"}}

	if err := store.Walkthroughs.UpsertWalkthrough(ctx, second); err != nil {
		t.Fatalf("second: %v", err)
	}

	got, _, _ := store.Walkthroughs.GetWalkthrough(ctx, "/r", "branch", "sha")

	if got.Summary != "v2" || got.Groups[0].Title != "v2" {
		t.Fatalf("expected overwrite to v2, got %#v", got)
	}
}
