package storage

import (
	"context"
	"path/filepath"
	"testing"
)

// newTestStore opens a fresh on-disk SQLite store rooted in t.TempDir() and
// runs every migration. Cleanup is registered automatically.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	return store
}

// seedUser ensures a user row exists and returns it. Distinct usernames
// produce distinct users; calling with the same name twice is idempotent
// and returns the same row (see UserRepo.EnsureUser).
func seedUser(t *testing.T, store *Store, osUsername string) User {
	t.Helper()

	user, err := store.Users.EnsureUser(context.Background(), osUsername)

	if err != nil {
		t.Fatalf("seed user %q: %v", osUsername, err)
	}

	return user
}

// seedRepo upserts a repository row owned by ownerID.
func seedRepo(t *testing.T, store *Store, ownerID int64, path, name string) Repository {
	t.Helper()

	repo, err := store.Repos.UpsertRepository(context.Background(), ownerID, path, name)

	if err != nil {
		t.Fatalf("seed repo %q: %v", path, err)
	}

	return repo
}

// seedBranches calls SyncBranches for the given names.
func seedBranches(t *testing.T, store *Store, repositoryID int64, names ...string) {
	t.Helper()

	if err := store.Branches.SyncBranches(context.Background(), repositoryID, names); err != nil {
		t.Fatalf("seed branches %v: %v", names, err)
	}
}

// seedReview creates a review session attached to userID with sensible
// defaults; the caller overrides any non-default field on the returned
// struct.
func seedReview(t *testing.T, store *Store, userID int64, start ReviewSessionStart) ReviewSession {
	t.Helper()

	if start.RepoRoot == "" {
		start.RepoRoot = "/tmp/repo"
	}

	if start.Branch == "" {
		start.Branch = "main"
	}

	if start.HeadSHA == "" {
		start.HeadSHA = "deadbeef"
	}

	if start.ContextKind == "" {
		start.ContextKind = "branch"
	}

	review, err := store.Reviews.CreateReview(context.Background(), userID, start)

	if err != nil {
		t.Fatalf("seed review: %v", err)
	}

	return review
}

// seedReviewComment inserts a comment row tied to reviewID.
func seedReviewComment(t *testing.T, store *Store, reviewID int64, input ReviewCommentInput) ReviewComment {
	t.Helper()

	if input.FilePath == "" {
		input.FilePath = "src/main.go"
	}

	if input.Side == "" {
		input.Side = "right"
	}

	if input.LineNumber == 0 {
		input.LineNumber = 1
	}

	if input.AuthorLabel == "" {
		input.AuthorLabel = "tester"
	}

	if input.BodyHTML == "" {
		input.BodyHTML = "<p>note</p>"
	}

	comment, err := store.Comments.CreateReviewComment(context.Background(), reviewID, input)

	if err != nil {
		t.Fatalf("seed comment: %v", err)
	}

	return comment
}

// seedPendingComment inserts a pending comment row owned by userID.
func seedPendingComment(t *testing.T, store *Store, userID int64, input PendingCommentInput) PendingComment {
	t.Helper()

	if input.RepoRoot == "" {
		input.RepoRoot = "/tmp/repo"
	}

	if input.ContextKind == "" {
		input.ContextKind = "branch"
	}

	if input.FilePath == "" {
		input.FilePath = "src/main.go"
	}

	if input.Side == "" {
		input.Side = "right"
	}

	if input.LineNumber == 0 {
		input.LineNumber = 1
	}

	if input.AuthorLabel == "" {
		input.AuthorLabel = "tester"
	}

	if input.BodyHTML == "" {
		input.BodyHTML = "<p>pending</p>"
	}

	pending, err := store.PendingComments.CreatePendingComment(context.Background(), userID, input)

	if err != nil {
		t.Fatalf("seed pending comment: %v", err)
	}

	return pending
}

// seedReviewEvent appends an event onto reviewID.
func seedReviewEvent(t *testing.T, store *Store, reviewID int64, input ReviewEventInput) ReviewEvent {
	t.Helper()

	if input.Type == "" {
		input.Type = "note"
	}

	event, err := store.ReviewEvents.Add(context.Background(), reviewID, input)

	if err != nil {
		t.Fatalf("seed review event: %v", err)
	}

	return event
}

// seedCollaborator grants role on repoPath to userID by ownerID.
func seedCollaborator(t *testing.T, store *Store, ownerID int64, repoPath string, userID int64, role string) RepositoryCollaborator {
	t.Helper()

	if role == "" {
		role = "reviewer"
	}

	collab, err := store.Collaborators.Grant(context.Background(), ownerID, repoPath, userID, role)

	if err != nil {
		t.Fatalf("seed collaborator: %v", err)
	}

	return collab
}

// seedWalkthrough upserts a walkthrough record.
func seedWalkthrough(t *testing.T, store *Store, record WalkthroughRecord) {
	t.Helper()

	if record.RepoRoot == "" {
		record.RepoRoot = "/tmp/repo"
	}

	if record.ContextKind == "" {
		record.ContextKind = "branch"
	}

	if record.ContextSHA == "" {
		record.ContextSHA = "deadbeef"
	}

	if record.ProviderID == "" {
		record.ProviderID = "anthropic"
	}

	if record.ModelID == "" {
		record.ModelID = "claude-sonnet-4-5"
	}

	if err := store.Walkthroughs.UpsertWalkthrough(context.Background(), record); err != nil {
		t.Fatalf("seed walkthrough: %v", err)
	}
}
