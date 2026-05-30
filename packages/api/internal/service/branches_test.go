package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newBranchService(t *testing.T, store *storage.Store) *BranchService {
	t.Helper()

	return NewBranchService(store.Branches, store.Repos)
}

func TestBranchServiceListSyncsRegisteredRepo(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	repoRoot := newGitRepo(t)

	if _, err := store.Repos.UpsertRepository(ctx, user.ID, repoRoot, "test-repo"); err != nil {
		t.Fatalf("register repo: %v", err)
	}

	svc := newBranchService(t, store)
	result, err := svc.List(ctx, repoRoot)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(result.Names) == 0 || result.Names[0] != "main" {
		t.Fatalf("expected at least one branch named 'main', got %v", result.Names)
	}

	if len(result.Records) == 0 {
		t.Fatalf("expected records populated for a registered repo, got %v", result.Records)
	}
}

func TestBranchServiceListUnregisteredRepoReturnsNamesOnly(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	repoRoot := newGitRepo(t)
	svc := newBranchService(t, store)

	result, err := svc.List(ctx, repoRoot)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(result.Names) == 0 {
		t.Fatalf("expected branch names, got %v", result.Names)
	}

	if len(result.Records) != 0 {
		t.Fatalf("expected zero records when repo unregistered, got %v", result.Records)
	}
}

func TestBranchServiceLockRequiresAuth(t *testing.T) {
	store := newTestStore(t)
	svc := newBranchService(t, store)

	if _, err := svc.Lock(context.Background(), 0, "/r", "main"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}
}

func TestBranchServiceLockReturnsNotFoundForUnregisteredRepo(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	repoRoot := newGitRepo(t)
	svc := newBranchService(t, store)

	if _, err := svc.Lock(ctx, user.ID, repoRoot, "main"); !errors.Is(err, storage.ErrRepositoryNotFound) {
		t.Fatalf("expected ErrRepositoryNotFound, got %v", err)
	}
}

func TestBranchServiceLockAndUnlockRegisteredRepo(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	repoRoot := newGitRepo(t)

	if _, err := store.Repos.UpsertRepository(ctx, user.ID, repoRoot, "test-repo"); err != nil {
		t.Fatalf("register: %v", err)
	}

	svc := newBranchService(t, store)

	// Populate branches.
	if _, err := svc.List(ctx, repoRoot); err != nil {
		t.Fatalf("list: %v", err)
	}

	records, err := svc.Lock(ctx, user.ID, repoRoot, "main")

	if err != nil {
		t.Fatalf("lock: %v", err)
	}

	foundLocked := false

	for _, r := range records {
		if r.Name == "main" && r.Locked {
			foundLocked = true
		}
	}

	if !foundLocked {
		t.Fatalf("expected main to be locked in returned records, got %v", records)
	}

	records, err = svc.Unlock(ctx, user.ID, repoRoot, "main")

	if err != nil {
		t.Fatalf("unlock: %v", err)
	}

	for _, r := range records {
		if r.Name == "main" && r.Locked {
			t.Fatalf("expected main unlocked, got %v", records)
		}
	}
}

func TestBranchServiceDeleteRequiresNameAndAuth(t *testing.T) {
	store := newTestStore(t)
	svc := newBranchService(t, store)

	if err := svc.Delete(context.Background(), 0, "/r", "x"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("expected ErrAuthenticationRequired, got %v", err)
	}

	user := seedUser(t, store, "alice")

	if err := svc.Delete(context.Background(), user.ID, "/r", ""); !errors.Is(err, ErrBranchNameRequired) {
		t.Fatalf("expected ErrBranchNameRequired, got %v", err)
	}
}

func TestBranchServiceDeleteRefusesLockedBranch(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	repoRoot := newGitRepo(t)

	if _, err := store.Repos.UpsertRepository(ctx, user.ID, repoRoot, "test-repo"); err != nil {
		t.Fatalf("register: %v", err)
	}

	svc := newBranchService(t, store)

	if _, err := svc.List(ctx, repoRoot); err != nil {
		t.Fatalf("list: %v", err)
	}

	if _, err := svc.Lock(ctx, user.ID, repoRoot, "main"); err != nil {
		t.Fatalf("lock: %v", err)
	}

	if err := svc.Delete(ctx, user.ID, repoRoot, "main"); !errors.Is(err, ErrBranchAlreadyLocked) {
		t.Fatalf("expected ErrBranchAlreadyLocked, got %v", err)
	}
}
