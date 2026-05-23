package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

// newTestStore mirrors the storage-layer helper: opens a fresh SQLite DB in
// t.TempDir(), runs migrations, and registers cleanup. Every service test
// builds its own store so concurrent tests can't share state.
func newTestStore(t *testing.T) *storage.Store {
	t.Helper()

	store, err := storage.Open(context.Background(), filepath.Join(t.TempDir(), "service.sqlite3"))

	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	return store
}

// seedUser ensures a user exists; equivalent to storage.seedUser but
// available inside the service package's tests.
func seedUser(t *testing.T, store *storage.Store, osUsername string) storage.User {
	t.Helper()

	user, err := store.Users.EnsureUser(context.Background(), osUsername)

	if err != nil {
		t.Fatalf("seed user %q: %v", osUsername, err)
	}

	return user
}

// seedRepo upserts a repository owned by ownerID at path.
func seedRepo(t *testing.T, store *storage.Store, ownerID int64, path, name string) storage.Repository {
	t.Helper()

	repo, err := store.Repos.UpsertRepository(context.Background(), ownerID, path, name)

	if err != nil {
		t.Fatalf("seed repo %q: %v", path, err)
	}

	return repo
}
