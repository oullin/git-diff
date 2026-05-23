package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestCreatePendingCommentDefaultsContextKind(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	pending, err := store.PendingComments.CreatePendingComment(ctx, user.ID, PendingCommentInput{
		RepoRoot:   "/r",
		FilePath:   "a.go",
		Side:       "right",
		LineNumber: 1,
		BodyHTML:   "<p>note</p>",
	})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if pending.ContextKind != "working" {
		t.Fatalf("expected default context kind 'working', got %q", pending.ContextKind)
	}
}

func TestCreatePendingCommentRequiresUserID(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.PendingComments.CreatePendingComment(context.Background(), 0, PendingCommentInput{}); err == nil {
		t.Fatalf("expected user id required")
	}
}

func TestUpdatePendingCommentRejectsForeignOwner(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	alice := seedUser(t, store, "alice")
	bob := seedUser(t, store, "bob")
	pending := seedPendingComment(t, store, alice.ID, PendingCommentInput{
		RepoRoot:   "/r",
		FilePath:   "a.go",
		LineNumber: 1,
	})

	if _, err := store.PendingComments.UpdatePendingComment(ctx, bob.ID, pending.ID, "<p>bob</p>"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows when updating another user's pending, got %v", err)
	}
}

func TestUpdatePendingCommentReplacesBody(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	pending := seedPendingComment(t, store, user.ID, PendingCommentInput{RepoRoot: "/r", FilePath: "a.go", LineNumber: 1, BodyHTML: "v1"})

	updated, err := store.PendingComments.UpdatePendingComment(ctx, user.ID, pending.ID, "v2")

	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.BodyHTML != "v2" {
		t.Fatalf("expected v2, got %q", updated.BodyHTML)
	}
}

func TestDeletePendingCommentRespectsOwnership(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	alice := seedUser(t, store, "alice")
	bob := seedUser(t, store, "bob")
	pending := seedPendingComment(t, store, alice.ID, PendingCommentInput{RepoRoot: "/r", FilePath: "a.go", LineNumber: 1})

	// Bob deleting Alice's pending must not actually delete the row.
	if err := store.PendingComments.DeletePendingComment(ctx, bob.ID, pending.ID); err != nil {
		t.Fatalf("foreign delete should not error: %v", err)
	}

	if _, err := store.PendingComments.GetPendingComment(ctx, alice.ID, pending.ID); err != nil {
		t.Fatalf("alice's pending must still exist, got %v", err)
	}

	if err := store.PendingComments.DeletePendingComment(ctx, alice.ID, pending.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
}

func TestListPendingCommentsFiltersByScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	seedPendingComment(t, store, user.ID, PendingCommentInput{
		RepoRoot:    "/r",
		ContextKind: "branch",
		ContextSHA:  "sha-A",
		FilePath:    "a.go",
		LineNumber:  1,
	})

	seedPendingComment(t, store, user.ID, PendingCommentInput{
		RepoRoot:    "/r",
		ContextKind: "branch",
		ContextSHA:  "sha-B",
		FilePath:    "b.go",
		LineNumber:  1,
	})

	list, err := store.PendingComments.ListPendingComments(ctx, user.ID, "/r", "branch", "sha-A")

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 in scope, got %d", len(list))
	}

	if list[0].ContextSHA != "sha-A" {
		t.Fatalf("expected sha-A, got %q", list[0].ContextSHA)
	}
}

func TestPromotePendingCommentsMovesRowsAndDeletesPending(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	review := seedReview(t, store, user.ID, ReviewSessionStart{RepoRoot: "/r", ContextKind: "branch", ContextSHA: "sha"})

	for i := 0; i < 3; i++ {
		seedPendingComment(t, store, user.ID, PendingCommentInput{
			RepoRoot:    "/r",
			ContextKind: "branch",
			ContextSHA:  "sha",
			FilePath:    "a.go",
			LineNumber:  int64(i + 1),
			BodyHTML:    "draft",
		})
	}

	count, err := store.PendingComments.PromotePendingComments(ctx, user.ID, review.ID)

	if err != nil {
		t.Fatalf("promote: %v", err)
	}

	if count != 3 {
		t.Fatalf("expected 3 promoted, got %d", count)
	}

	pending, _ := store.PendingComments.ListPendingComments(ctx, user.ID, "/r", "branch", "sha")

	if len(pending) != 0 {
		t.Fatalf("expected pendings drained, got %v", pending)
	}

	comments, _ := store.Comments.ListReviewComments(ctx, review.ID)

	if len(comments) != 3 {
		t.Fatalf("expected 3 promoted comments, got %d", len(comments))
	}
}

func TestPromotePendingCommentsErrorsOnMissingReview(t *testing.T) {
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	if _, err := store.PendingComments.PromotePendingComments(context.Background(), user.ID, 9999); err == nil {
		t.Fatalf("expected error for missing review")
	}
}
