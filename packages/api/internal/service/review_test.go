package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newReviewService(t *testing.T, store *storage.Store) *ReviewService {
	t.Helper()

	return NewReviewService(store.Reviews, store.ReviewEvents, store.Comments)
}

func TestReviewServiceCreateRequiresUser(t *testing.T) {
	store := newTestStore(t)
	svc := newReviewService(t, store)

	if _, err := svc.Create(context.Background(), 0, storage.ReviewSessionStart{RepoRoot: "/r"}); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestReviewServiceCreateRoundTrip(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newReviewService(t, store)

	created, err := svc.Create(ctx, user.ID, storage.ReviewSessionStart{RepoRoot: "/r", HeadSHA: "abc"})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected an id assigned")
	}
}

func TestReviewServiceListDefaultsLimit(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newReviewService(t, store)

	if _, err := svc.Create(ctx, user.ID, storage.ReviewSessionStart{RepoRoot: "/r"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := svc.List(ctx, user.ID, -5)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) == 0 {
		t.Fatalf("expected at least one review")
	}
}

func TestReviewServiceListRequiresUser(t *testing.T) {
	store := newTestStore(t)
	svc := newReviewService(t, store)

	if _, err := svc.List(context.Background(), 0, 10); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestReviewServiceDetailHasEventsAndComments(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newReviewService(t, store)

	created, err := svc.Create(ctx, user.ID, storage.ReviewSessionStart{RepoRoot: "/r"})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := svc.AddEvent(ctx, created.ID, storage.ReviewEventInput{Type: "viewed"}); err != nil {
		t.Fatalf("add event: %v", err)
	}

	if _, err := svc.CreateComment(ctx, created.ID, storage.ReviewCommentInput{FilePath: "a.go", LineNumber: 1, BodyHTML: "x"}); err != nil {
		t.Fatalf("create comment: %v", err)
	}

	detail, err := svc.Detail(ctx, created.ID)

	if err != nil {
		t.Fatalf("detail: %v", err)
	}

	if len(detail.Events) < 2 {
		t.Fatalf("expected at least review_started + viewed events, got %d", len(detail.Events))
	}

	if len(detail.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(detail.Comments))
	}
}

func TestReviewServiceUpdateAndDeleteComment(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newReviewService(t, store)

	review, err := svc.Create(ctx, user.ID, storage.ReviewSessionStart{RepoRoot: "/r"})

	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	comment, err := svc.CreateComment(ctx, review.ID, storage.ReviewCommentInput{
		FilePath: "a.go", LineNumber: 1, BodyHTML: "<p>v1</p>",
	})

	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	updated, err := svc.UpdateComment(ctx, review.ID, comment.ID, "<p>v2</p>")

	if err != nil {
		t.Fatalf("update comment: %v", err)
	}

	if updated.BodyHTML != "<p>v2</p>" {
		t.Fatalf("expected updated body, got %q", updated.BodyHTML)
	}

	if err := svc.DeleteComment(ctx, review.ID, comment.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	detail, _ := svc.Detail(ctx, review.ID)

	if len(detail.Comments) != 0 {
		t.Fatalf("expected soft-deleted comment to be filtered out, got %v", detail.Comments)
	}
}
