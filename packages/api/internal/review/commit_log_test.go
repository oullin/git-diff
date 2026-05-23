package review

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListCommitLogReturnsCommits(t *testing.T) {
	repo := newGitRepo(t)

	commits, err := ListCommitLog(context.Background(), repo, 0)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(commits) == 0 {
		t.Fatalf("expected at least one commit")
	}

	if commits[0].SHA == "" || commits[0].ShortSHA == "" {
		t.Fatalf("expected SHA fields populated, got %+v", commits[0])
	}

	if commits[0].Subject != "initial" {
		t.Fatalf("expected initial subject, got %q", commits[0].Subject)
	}
}

func TestListCommitLogHonoursLimit(t *testing.T) {
	repo := newGitRepo(t)

	// Add two more commits.
	for i := 0; i < 2; i++ {
		file := filepath.Join(repo, "f"+string(rune('A'+i))+".txt")

		if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		runGit(t, "-C", repo, "add", "-A")
		runGit(t, "-C", repo, "commit", "-m", "extra")
	}

	commits, err := ListCommitLog(context.Background(), repo, 2)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected limit=2 commits, got %d", len(commits))
	}
}

func TestListCommitLogClampsExtremeLimit(t *testing.T) {
	repo := newGitRepo(t)

	// limit < 0 and > 500 both fall back to default of 100.
	if _, err := ListCommitLog(context.Background(), repo, -10); err != nil {
		t.Fatalf("negative limit should not error, got %v", err)
	}

	if _, err := ListCommitLog(context.Background(), repo, 10_000); err != nil {
		t.Fatalf("oversize limit should not error, got %v", err)
	}
}
