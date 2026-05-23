package storage

import (
	"context"
	"testing"
)

func TestCreateReviewDefaultsAndCreatesStartEvent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	// All optional fields empty: defaults must fill in Title and ContextKind.
	review, err := store.Reviews.CreateReview(ctx, user.ID, ReviewSessionStart{
		RepoRoot: "/repo",
		Branch:   "main",
		HeadSHA:  "abc",
	})

	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	if review.Title != "Local review" {
		t.Fatalf("expected default title, got %q", review.Title)
	}

	if review.ContextKind != "working" {
		t.Fatalf("expected default context kind, got %q", review.ContextKind)
	}

	events, err := store.ReviewEvents.List(ctx, review.ID)

	if err != nil {
		t.Fatalf("list events: %v", err)
	}

	if len(events) != 1 || events[0].Type != "review_started" {
		t.Fatalf("expected single review_started event, got %v", events)
	}
}

func TestCreateReviewRequiresUserID(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Reviews.CreateReview(context.Background(), 0, ReviewSessionStart{RepoRoot: "/r"}); err == nil {
		t.Fatalf("expected user id required")
	}
}

func TestListReviewsHonoursLimitAndOrder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	for i := 0; i < 5; i++ {
		seedReview(t, store, user.ID, ReviewSessionStart{
			RepoRoot: "/repo",
			HeadSHA:  "sha" + string(rune('A'+i)),
		})
	}

	list, err := store.Reviews.ListReviews(ctx, user.ID, 3)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected limit=3, got %d", len(list))
	}

	// Result order is by started_at DESC; with our seed loop the last one
	// inserted has the highest started_at, so it must come first.
	if list[0].HeadSHA == "" || list[0].HeadSHA != "shaE" {
		t.Fatalf("expected newest review first, got %q", list[0].HeadSHA)
	}
}

func TestListReviewsDefaultLimitOnZero(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	list, err := store.Reviews.ListReviews(ctx, user.ID, 0)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) == 0 {
		t.Fatalf("expected at least one review")
	}
}

func TestGetReviewByIDReturnsErrorWhenMissing(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Reviews.GetReviewByID(context.Background(), 9999); err == nil {
		t.Fatalf("expected error for missing review")
	}
}

func TestReviewDetailAggregatesEventsAndComments(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	seedReviewEvent(t, store, review.ID, ReviewEventInput{Type: "viewed", FilePath: "a.go"})
	seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 10})

	detail, err := store.Reviews.ReviewDetail(ctx, store.Comments, review.ID)

	if err != nil {
		t.Fatalf("detail: %v", err)
	}

	if detail.Review.ID != review.ID {
		t.Fatalf("expected review id %d, got %d", review.ID, detail.Review.ID)
	}

	// Events include the implicit review_started + the explicit viewed + the
	// implicit comment_added emitted by CreateReviewComment = 3.
	if len(detail.Events) != 3 {
		t.Fatalf("expected 3 events, got %d (%v)", len(detail.Events), detail.Events)
	}

	if len(detail.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(detail.Comments))
	}
}
