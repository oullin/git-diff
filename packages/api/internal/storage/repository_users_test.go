package storage

import (
	"context"
	"errors"
	"testing"
)

func TestCollaboratorGrantAndList(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	bob := seedUser(t, store, "bob")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if _, err := store.Collaborators.Grant(ctx, owner.ID, repo.Path, alice.ID, RepoRoleWrite); err != nil {
		t.Fatalf("grant alice: %v", err)
	}

	if _, err := store.Collaborators.Grant(ctx, owner.ID, repo.Path, bob.ID, RepoRoleRead); err != nil {
		t.Fatalf("grant bob: %v", err)
	}

	list, err := store.Collaborators.List(ctx, owner.ID, repo.Path)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 collaborators, got %d (%v)", len(list), list)
	}

	if list[0].OSUsername != "alice" || list[1].OSUsername != "bob" {
		t.Fatalf("expected alphabetical ordering, got %q, %q", list[0].OSUsername, list[1].OSUsername)
	}
}

func TestCollaboratorGrantUpsertsRoleOnConflict(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if _, err := store.Collaborators.Grant(ctx, owner.ID, repo.Path, alice.ID, RepoRoleRead); err != nil {
		t.Fatalf("first grant: %v", err)
	}

	updated, err := store.Collaborators.Grant(ctx, owner.ID, repo.Path, alice.ID, RepoRoleWrite)

	if err != nil {
		t.Fatalf("second grant: %v", err)
	}

	if updated.Role != RepoRoleWrite {
		t.Fatalf("expected role updated to write, got %q", updated.Role)
	}
}

func TestCollaboratorGrantRejectsInvalidRole(t *testing.T) {
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if _, err := store.Collaborators.Grant(context.Background(), owner.ID, repo.Path, alice.ID, "admin"); err == nil {
		t.Fatalf("expected error for invalid role")
	}
}

func TestCollaboratorGrantRejectsOwnerAsCollaborator(t *testing.T) {
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if _, err := store.Collaborators.Grant(context.Background(), owner.ID, repo.Path, owner.ID, RepoRoleWrite); err == nil {
		t.Fatalf("expected error for owner-as-collaborator")
	}
}

func TestCollaboratorGrantDeniesNonOwner(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	intruder := seedUser(t, store, "intruder")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if _, err := store.Collaborators.Grant(ctx, intruder.ID, repo.Path, alice.ID, RepoRoleRead); !errors.Is(err, ErrRepositoryNotOwned) {
		t.Fatalf("expected ErrRepositoryNotOwned, got %v", err)
	}
}

func TestCollaboratorGrantReturnsNotFoundForMissingRepo(t *testing.T) {
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")

	if _, err := store.Collaborators.Grant(context.Background(), owner.ID, "/never", alice.ID, RepoRoleRead); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("expected ErrRepositoryNotFound, got %v", err)
	}
}

func TestCollaboratorRevokeRemovesRow(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	alice := seedUser(t, store, "alice")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	seedCollaborator(t, store, owner.ID, repo.Path, alice.ID, RepoRoleRead)

	if err := store.Collaborators.Revoke(ctx, owner.ID, repo.Path, alice.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	list, _ := store.Collaborators.List(ctx, owner.ID, repo.Path)

	if len(list) != 0 {
		t.Fatalf("expected zero collaborators after revoke, got %v", list)
	}
}

func TestCollaboratorListEmptyReturnsNonNilSlice(t *testing.T) {
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	list, err := store.Collaborators.List(context.Background(), owner.ID, repo.Path)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if list == nil {
		t.Fatalf("expected non-nil empty slice")
	}

	if len(list) != 0 {
		t.Fatalf("expected empty list, got %v", list)
	}
}
