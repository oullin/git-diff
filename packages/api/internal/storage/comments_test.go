package storage

import (
	"context"
	"testing"
)

func TestCreateReviewCommentDefaultsAuthorLabel(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	comment, err := store.Comments.CreateReviewComment(ctx, review.ID, ReviewCommentInput{
		FilePath:   "a.go",
		Side:       "right",
		LineNumber: 1,
		BodyHTML:   "<p>note</p>",
	})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if comment.AuthorLabel != "You" {
		t.Fatalf("expected default author label, got %q", comment.AuthorLabel)
	}
}

func TestCreateReviewCommentEmitsCommentAddedEvent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 42})

	events, _ := store.ReviewEvents.List(ctx, review.ID)

	found := false

	for _, e := range events {
		if e.Type == "comment_added" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected comment_added event after CreateReviewComment, got %v", events)
	}
}

func TestUpdateReviewCommentUndeletesAndEmitsEditEvent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})
	comment := seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 1, BodyHTML: "v1"})

	if err := store.Comments.DeleteReviewComment(ctx, review.ID, comment.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	updated, err := store.Comments.UpdateReviewComment(ctx, review.ID, comment.ID, "v2")

	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.BodyHTML != "v2" {
		t.Fatalf("expected body to update to v2, got %q", updated.BodyHTML)
	}

	if updated.DeletedAt != "" {
		t.Fatalf("expected deleted_at cleared after update, got %q", updated.DeletedAt)
	}

	events, _ := store.ReviewEvents.List(ctx, review.ID)

	found := false

	for _, e := range events {
		if e.Type == "comment_edited" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected comment_edited event, got %v", events)
	}
}

func TestDeleteReviewCommentIsSoftDelete(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})
	comment := seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 1})

	if err := store.Comments.DeleteReviewComment(ctx, review.ID, comment.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// ListReviewComments must exclude soft-deleted rows.
	list, _ := store.Comments.ListReviewComments(ctx, review.ID)

	if len(list) != 0 {
		t.Fatalf("expected list to exclude deleted, got %v", list)
	}

	// But the row still exists with deleted_at set.
	row, err := store.Comments.GetReviewComment(ctx, review.ID, comment.ID)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if row.DeletedAt == "" {
		t.Fatalf("expected deleted_at set after delete")
	}
}

func TestSetReviewCommentResolvedTogglesAndEmitsEvent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})
	comment := seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 1})

	resolved, err := store.Comments.SetReviewCommentResolved(ctx, review.ID, comment.ID, true)

	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if !resolved.Resolved {
		t.Fatalf("expected comment resolved")
	}

	if resolved.ResolvedAt == "" {
		t.Fatalf("expected resolved_at set after resolve")
	}

	reopened, err := store.Comments.SetReviewCommentResolved(ctx, review.ID, comment.ID, false)

	if err != nil {
		t.Fatalf("unresolve: %v", err)
	}

	if reopened.Resolved {
		t.Fatalf("expected comment unresolved")
	}

	if reopened.ResolvedAt != "" {
		t.Fatalf("expected resolved_at cleared after unresolve, got %q", reopened.ResolvedAt)
	}

	events, _ := store.ReviewEvents.List(ctx, review.ID)

	foundResolved := false
	foundUnresolved := false

	for _, e := range events {
		switch e.Type {
		case "comment_resolved":
			foundResolved = true
		case "comment_unresolved":
			foundUnresolved = true
		}
	}

	if !foundResolved || !foundUnresolved {
		t.Fatalf("expected resolve and unresolve events, got %v", events)
	}
}

func TestListReviewCommentsIncludesResolved(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})
	comment := seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 1})

	if _, err := store.Comments.SetReviewCommentResolved(ctx, review.ID, comment.ID, true); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	list, err := store.Comments.ListReviewComments(ctx, review.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected resolved comment still listed, got %d", len(list))
	}

	if !list[0].Resolved {
		t.Fatalf("expected listed comment to be resolved")
	}
}

func TestListReviewCommentsOrderedByCreatedAt(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r"})

	seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "a.go", LineNumber: 1, BodyHTML: "first"})
	seedReviewComment(t, store, review.ID, ReviewCommentInput{FilePath: "b.go", LineNumber: 1, BodyHTML: "second"})

	list, err := store.Comments.ListReviewComments(ctx, review.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(list))
	}

	if list[0].BodyHTML != "first" {
		t.Fatalf("expected created_at ASC, got %q first", list[0].BodyHTML)
	}
}
