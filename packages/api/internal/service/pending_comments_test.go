package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newPendingService(t *testing.T, store *storage.Store) *PendingCommentService {
	t.Helper()

	return NewPendingCommentService(store.PendingComments)
}

func TestPendingCommentServiceRequiresAuth(t *testing.T) {
	store := newTestStore(t)
	svc := newPendingService(t, store)
	ctx := context.Background()

	if _, err := svc.List(ctx, 0, "/r", "branch", ""); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("List should require auth, got %v", err)
	}

	if _, err := svc.Create(ctx, 0, "/r", storage.PendingCommentInput{}); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Create should require auth, got %v", err)
	}

	if _, err := svc.Update(ctx, 0, 1, "x"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Update should require auth, got %v", err)
	}

	if err := svc.Delete(ctx, 0, 1); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Delete should require auth, got %v", err)
	}

	if _, err := svc.Promote(ctx, 0, 1); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Promote should require auth, got %v", err)
	}
}

func TestPendingCommentServiceCreateValidatesRequiredFields(t *testing.T) {
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newPendingService(t, store)
	ctx := context.Background()

	if _, err := svc.Create(ctx, user.ID, "/r", storage.PendingCommentInput{}); !errors.Is(err, ErrInvalidCommentInput) {
		t.Fatalf("expected ErrInvalidCommentInput when filePath/diffSection blank, got %v", err)
	}
}

func TestPendingCommentServiceCreateAppliesDefaultRepoRoot(t *testing.T) {
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newPendingService(t, store)

	pending, err := svc.Create(context.Background(), user.ID, "/default-repo", storage.PendingCommentInput{
		FilePath:    "a.go",
		DiffSection: "diff",
		Side:        "right",
		LineNumber:  1,
		BodyHTML:    "<p>note</p>",
	})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if pending.RepoRoot != "/default-repo" {
		t.Fatalf("expected default repo root, got %q", pending.RepoRoot)
	}
}

func TestPendingCommentServicePromoteRequiresReviewID(t *testing.T) {
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newPendingService(t, store)

	if _, err := svc.Promote(context.Background(), user.ID, 0); !errors.Is(err, ErrReviewIDRequired) {
		t.Fatalf("expected ErrReviewIDRequired, got %v", err)
	}
}

func TestPendingCommentServiceListThroughDelete(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newPendingService(t, store)

	created, err := svc.Create(ctx, user.ID, "/r", storage.PendingCommentInput{
		FilePath:    "a.go",
		DiffSection: "diff",
		Side:        "right",
		LineNumber:  1,
		BodyHTML:    "<p>v1</p>",
	})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := svc.List(ctx, user.ID, "/r", "working", "")

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(list))
	}

	if _, err := svc.Update(ctx, user.ID, created.ID, "<p>v2</p>"); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := svc.Delete(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
