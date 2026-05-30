package storage

import (
	"context"
	"testing"
)

func TestAddReviewEventDefaultsMetadataToEmptyObject(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	event, err := store.ReviewEvents.Add(ctx, review.ID, ReviewEventInput{Type: "viewed", FilePath: "a.go"})

	if err != nil {
		t.Fatalf("add: %v", err)
	}

	if event.Metadata != "{}" {
		t.Fatalf("expected default metadata {}, got %q", event.Metadata)
	}
}

func TestAddReviewEventPreservesSuppliedMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	event, err := store.ReviewEvents.Add(ctx, review.ID, ReviewEventInput{
		Type:     "viewed",
		Metadata: `{"k":"v"}`,
	})

	if err != nil {
		t.Fatalf("add: %v", err)
	}

	if event.Metadata != `{"k":"v"}` {
		t.Fatalf("metadata not preserved, got %q", event.Metadata)
	}
}

func TestListReviewEventsOrderedAscending(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	seedReviewEvent(t, store, review.ID, ReviewEventInput{Type: "first"})
	seedReviewEvent(t, store, review.ID, ReviewEventInput{Type: "second"})

	list, err := store.ReviewEvents.List(ctx, review.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(list))
	}

	// Find our seeded events; the first one inserted must come before the
	// second per the created_at ASC, id ASC ordering.
	var firstIdx, secondIdx = -1, -1

	for i, e := range list {
		if e.Type == "first" && firstIdx == -1 {
			firstIdx = i
		}

		if e.Type == "second" && secondIdx == -1 {
			secondIdx = i
		}
	}

	if firstIdx == -1 || secondIdx == -1 || firstIdx > secondIdx {
		t.Fatalf("expected first before second; got positions %d, %d", firstIdx, secondIdx)
	}
}
