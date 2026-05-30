package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newCollaboratorService(t *testing.T, store *storage.Store) *CollaboratorService {
	t.Helper()

	return NewCollaboratorService(store.Collaborators)
}

func TestCollaboratorServiceRequiresAuthOnList(t *testing.T) {
	store := newTestStore(t)
	svc := newCollaboratorService(t, store)

	if _, err := svc.List(context.Background(), 0, "/r"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestCollaboratorServiceRequiresAuthOnGrant(t *testing.T) {
	store := newTestStore(t)
	svc := newCollaboratorService(t, store)

	if _, err := svc.Grant(context.Background(), 0, "/r", 1, storage.RepoRoleRead); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestCollaboratorServiceRequiresAuthOnRevoke(t *testing.T) {
	store := newTestStore(t)
	svc := newCollaboratorService(t, store)

	if err := svc.Revoke(context.Background(), 0, "/r", 1); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestCollaboratorServiceGrantAndList(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	svc := newCollaboratorService(t, store)

	if _, err := svc.Grant(ctx, owner.ID, repo.Path, alice.ID, storage.RepoRoleWrite); err != nil {
		t.Fatalf("grant: %v", err)
	}

	list, err := svc.List(ctx, owner.ID, repo.Path)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 || list[0].UserID != alice.ID {
		t.Fatalf("expected alice as collaborator, got %v", list)
	}
}

func TestCollaboratorServiceRevoke(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	svc := newCollaboratorService(t, store)

	if _, err := svc.Grant(ctx, owner.ID, repo.Path, alice.ID, storage.RepoRoleWrite); err != nil {
		t.Fatalf("grant: %v", err)
	}

	if err := svc.Revoke(ctx, owner.ID, repo.Path, alice.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	list, _ := svc.List(ctx, owner.ID, repo.Path)

	if len(list) != 0 {
		t.Fatalf("expected list to be empty after revoke, got %v", list)
	}
}
