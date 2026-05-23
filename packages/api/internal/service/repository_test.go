package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newRepositoryService(t *testing.T, store *storage.Store) *RepositoryService {
	t.Helper()

	return NewRepositoryService(store.Repos)
}

func TestRepositoryServiceRequiresAuth(t *testing.T) {
	store := newTestStore(t)
	svc := newRepositoryService(t, store)

	if _, err := svc.List(context.Background(), 0); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("List should require auth, got %v", err)
	}

	if _, err := svc.Upsert(context.Background(), 0, "/r", "r"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Upsert should require auth, got %v", err)
	}

	if err := svc.Remove(context.Background(), 0, "/r"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Remove should require auth, got %v", err)
	}
}

func TestRepositoryServiceRemoveRejectsBlankPath(t *testing.T) {
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newRepositoryService(t, store)

	if err := svc.Remove(context.Background(), user.ID, ""); !errors.Is(err, ErrRepositoryPathRequired) {
		t.Fatalf("expected ErrRepositoryPathRequired, got %v", err)
	}
}

func TestRepositoryServiceUpsertListAndRemove(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newRepositoryService(t, store)

	repo, err := svc.Upsert(ctx, user.ID, "/r", "name")

	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if repo.OwnerID != user.ID {
		t.Fatalf("expected owner id %d, got %d", user.ID, repo.OwnerID)
	}

	list, err := svc.List(ctx, user.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(list))
	}

	if err := svc.Remove(ctx, user.ID, repo.Path); err != nil {
		t.Fatalf("remove: %v", err)
	}

	list, _ = svc.List(ctx, user.ID)

	if len(list) != 0 {
		t.Fatalf("expected empty after remove, got %v", list)
	}
}

func TestRepositoryServiceRemovePropagatesNotOwned(t *testing.T) {
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	intruder := seedUser(t, store, "intruder")
	seedRepo(t, store, owner.ID, "/r", "r")
	svc := newRepositoryService(t, store)

	if err := svc.Remove(context.Background(), intruder.ID, "/r"); !errors.Is(err, storage.ErrRepositoryNotOwned) {
		t.Fatalf("expected storage.ErrRepositoryNotOwned, got %v", err)
	}
}
