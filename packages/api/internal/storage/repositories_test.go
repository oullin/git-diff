package storage

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestUpsertRepositoryDefaultsNameToBaseDir(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")

	repo, err := store.Repos.UpsertRepository(ctx, owner.ID, "/Users/x/projects/sample", "")

	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if repo.Name != "sample" {
		t.Fatalf("expected default name from base dir, got %q", repo.Name)
	}
}

func TestUpsertRepositoryUpdatesNameOnConflict(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")

	first, err := store.Repos.UpsertRepository(ctx, owner.ID, "/repo", "first")

	if err != nil {
		t.Fatalf("upsert first: %v", err)
	}

	second, err := store.Repos.UpsertRepository(ctx, owner.ID, "/repo", "second")

	if err != nil {
		t.Fatalf("upsert second: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same id on conflict, got %d vs %d", first.ID, second.ID)
	}

	if second.Name != "second" {
		t.Fatalf("expected name to update to %q, got %q", "second", second.Name)
	}
}

func TestUpsertRepositoryRequiresOwnerAndPath(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Repos.UpsertRepository(context.Background(), 0, "/x", "x"); err == nil {
		t.Fatalf("expected owner id required")
	}

	owner := seedUser(t, store, "owner")

	if _, err := store.Repos.UpsertRepository(context.Background(), owner.ID, "", "x"); err == nil {
		t.Fatalf("expected path required")
	}
}

func TestListRepositoriesForUserIncludesOwnerAndCollaborator(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	other := seedUser(t, store, "other")

	ownerRepo := seedRepo(t, store, owner.ID, "/repo-owned", "owned")
	sharedRepo := seedRepo(t, store, other.ID, "/repo-shared", "shared")
	_ = seedRepo(t, store, other.ID, "/repo-unrelated", "unrelated")

	// Make owner a write collaborator on the shared repo.
	if _, err := store.Collaborators.Grant(ctx, other.ID, sharedRepo.Path, owner.ID, RepoRoleWrite); err != nil {
		t.Fatalf("grant: %v", err)
	}

	list, err := store.Repos.ListRepositoriesForUser(ctx, owner.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 repositories, got %d (%v)", len(list), list)
	}

	roles := map[string]string{}

	for _, r := range list {
		roles[r.Path] = r.Role
	}

	if roles[ownerRepo.Path] != RepoRoleOwner {
		t.Fatalf("expected owner role on %s, got %q", ownerRepo.Path, roles[ownerRepo.Path])
	}

	if roles[sharedRepo.Path] != RepoRoleWrite {
		t.Fatalf("expected write role on %s, got %q", sharedRepo.Path, roles[sharedRepo.Path])
	}
}

func TestGetRepositoryDeniesUnrelatedUsers(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	other := seedUser(t, store, "other")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	_, err := store.Repos.GetRepository(ctx, other.ID, repo.Path)

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected not found for unrelated user, got %v", err)
	}
}

func TestGetByPathDoesNotJoinRole(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	seeded := seedRepo(t, store, owner.ID, "/repo", "repo")

	got, err := store.Repos.GetByPath(ctx, seeded.Path)

	if err != nil {
		t.Fatalf("get by path: %v", err)
	}

	if got.ID != seeded.ID {
		t.Fatalf("expected id %d got %d", seeded.ID, got.ID)
	}

	if got.Role != "" {
		t.Fatalf("GetByPath must not populate Role, got %q", got.Role)
	}
}

func TestRemoveRepositoryEnforcesOwnership(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	owner := seedUser(t, store, "owner")
	other := seedUser(t, store, "other")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if err := store.Repos.RemoveRepository(ctx, other.ID, repo.Path); !errors.Is(err, ErrRepositoryNotOwned) {
		t.Fatalf("non-owner remove must return ErrRepositoryNotOwned, got %v", err)
	}

	if err := store.Repos.RemoveRepository(ctx, owner.ID, repo.Path); err != nil {
		t.Fatalf("owner remove failed: %v", err)
	}

	// Second remove for the same path must report ErrRepositoryNotOwned since
	// no row exists at all.
	if err := store.Repos.RemoveRepository(ctx, owner.ID, repo.Path); !errors.Is(err, ErrRepositoryNotOwned) {
		t.Fatalf("removing non-existent path must return ErrRepositoryNotOwned, got %v", err)
	}
}
