package storage

import (
	"context"
	"testing"
)

func TestSyncBranchesInsertsAndReconciles(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if err := store.Branches.SyncBranches(ctx, repo.ID, []string{"main", "feature/a", "feature/b"}); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	branches, err := store.Branches.ListBranches(ctx, repo.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(branches) != 3 {
		t.Fatalf("expected 3 branches, got %d (%v)", len(branches), branches)
	}

	// Second sync drops feature/b — but only because it isn't locked.
	if err := store.Branches.SyncBranches(ctx, repo.ID, []string{"main", "feature/a"}); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	branches, _ = store.Branches.ListBranches(ctx, repo.ID)

	names := map[string]bool{}

	for _, b := range branches {
		names[b.Name] = true
	}

	if names["feature/b"] {
		t.Fatalf("feature/b should have been removed by reconcile")
	}

	if !names["main"] || !names["feature/a"] {
		t.Fatalf("expected main + feature/a to remain, got %v", names)
	}
}

func TestSyncBranchesSkipsEmptyNames(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if err := store.Branches.SyncBranches(ctx, repo.ID, []string{"", "main", ""}); err != nil {
		t.Fatalf("sync: %v", err)
	}

	branches, _ := store.Branches.ListBranches(ctx, repo.ID)

	if len(branches) != 1 || branches[0].Name != "main" {
		t.Fatalf("expected only main, got %v", branches)
	}
}

func TestSyncBranchesRejectsZeroRepoID(t *testing.T) {
	store := newTestStore(t)

	if err := store.Branches.SyncBranches(context.Background(), 0, []string{"main"}); err == nil {
		t.Fatalf("expected error for zero repo id")
	}
}

func TestLockBranchSetsRowAndPreservesAcrossReconcile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	seedBranches(t, store, repo.ID, "main", "feature/a")

	if err := store.Branches.LockBranch(ctx, repo.ID, "feature/a", owner.ID); err != nil {
		t.Fatalf("lock: %v", err)
	}

	locked, err := store.Branches.IsBranchLocked(ctx, repo.ID, "feature/a")

	if err != nil || !locked {
		t.Fatalf("expected locked=true, got %v err=%v", locked, err)
	}

	// A reconcile that omits feature/a must NOT delete it because it's locked.
	if err := store.Branches.SyncBranches(ctx, repo.ID, []string{"main"}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	branches, _ := store.Branches.ListBranches(ctx, repo.ID)

	names := map[string]bool{}

	for _, b := range branches {
		names[b.Name] = true
	}

	if !names["feature/a"] {
		t.Fatalf("locked branch must survive reconcile; got %v", names)
	}
}

func TestLockBranchReturnsErrorOnUnknownBranch(t *testing.T) {
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	if err := store.Branches.LockBranch(context.Background(), repo.ID, "no-such", owner.ID); err == nil {
		t.Fatalf("expected error for missing branch")
	}
}

func TestUnlockBranchClearsLockMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	seedBranches(t, store, repo.ID, "main")

	if err := store.Branches.LockBranch(ctx, repo.ID, "main", owner.ID); err != nil {
		t.Fatalf("lock: %v", err)
	}

	if err := store.Branches.UnlockBranch(ctx, repo.ID, "main"); err != nil {
		t.Fatalf("unlock: %v", err)
	}

	branches, _ := store.Branches.ListBranches(ctx, repo.ID)

	for _, b := range branches {
		if b.Name != "main" {
			continue
		}

		if b.Locked || b.LockedAt != "" || b.LockedBy != 0 {
			t.Fatalf("expected lock cleared, got %#v", b)
		}
	}
}

func TestIsBranchLockedReturnsFalseForMissingBranch(t *testing.T) {
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")

	locked, err := store.Branches.IsBranchLocked(context.Background(), repo.ID, "ghost")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if locked {
		t.Fatalf("expected false for missing branch")
	}
}

func TestDeleteBranchRowRemovesRow(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	owner := seedUser(t, store, "owner")
	repo := seedRepo(t, store, owner.ID, "/repo", "repo")
	seedBranches(t, store, repo.ID, "main", "feature/a")

	if err := store.Branches.DeleteBranchRow(ctx, repo.ID, "feature/a"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	branches, _ := store.Branches.ListBranches(ctx, repo.ID)

	for _, b := range branches {
		if b.Name == "feature/a" {
			t.Fatalf("feature/a should be gone")
		}
	}
}
