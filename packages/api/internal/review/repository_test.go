package review

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRepositoryState(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")

	writeFile(t, root, "app.txt", "old\n")
	git(t, root, "add", "app.txt")
	git(t, root, "commit", "-m", "initial")

	writeFile(t, root, "app.txt", "old\nnew\n")
	writeFile(t, root, "notes.txt", "hello\n")

	state, err := ReadRepositoryState(context.Background(), root)

	if err != nil {
		t.Fatal(err)
	}

	wantRoot, err := filepath.EvalSymlinks(root)

	if err != nil {
		t.Fatal(err)
	}

	if state.Root != wantRoot {
		t.Fatalf("root = %q, want %q", state.Root, wantRoot)
	}

	if len(state.Files) != 2 {
		t.Fatalf("files = %d, want 2", len(state.Files))
	}

	if state.Additions == 0 {
		t.Fatal("expected additions")
	}
}

func TestReadCommitState(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")
	git(t, root, "config", "commit.gpgsign", "false")

	writeFile(t, root, "app.txt", "old\n")
	git(t, root, "add", "app.txt")
	git(t, root, "commit", "-m", "initial commit")

	writeFile(t, root, "app.txt", "old\nnew\n")
	writeFile(t, root, "second.txt", "hello\n")
	git(t, root, "add", "app.txt", "second.txt")
	git(t, root, "commit", "-m", "second commit")

	head := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD"))
	state, err := ReadCommitState(context.Background(), root, head)

	if err != nil {
		t.Fatal(err)
	}

	if state.Mode != RepositoryModeCommit {
		t.Fatalf("mode = %q, want %q", state.Mode, RepositoryModeCommit)
	}

	if state.CommitSHA != head {
		t.Fatalf("commit sha = %q, want %q", state.CommitSHA, head)
	}

	if len(state.Files) != 2 {
		t.Fatalf("files = %d, want 2", len(state.Files))
	}

	for _, file := range state.Files {
		if len(file.Sections) != 1 || file.Sections[0].Kind != "commit" {
			t.Fatalf("%s sections = %#v", file.Path, file.Sections)
		}
	}

	short := strings.TrimSpace(gitOut(t, root, "rev-parse", "--short", head))

	if _, err := ReadCommitState(context.Background(), root, short); err != nil {
		t.Fatalf("short sha lookup failed: %v", err)
	}

	if _, err := ReadCommitState(context.Background(), root, "deadbeef"); err == nil {
		t.Fatal("expected unknown commit to fail")
	}
}

func TestListCommitLog(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")
	git(t, root, "config", "commit.gpgsign", "false")

	writeFile(t, root, "a.txt", "1\n")
	git(t, root, "add", "a.txt")
	git(t, root, "commit", "-m", "first")
	writeFile(t, root, "a.txt", "2\n")
	git(t, root, "add", "a.txt")
	git(t, root, "commit", "-m", "second")

	commits, err := ListCommitLog(context.Background(), root, 10)

	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 2 {
		t.Fatalf("commits = %d, want 2", len(commits))
	}

	if commits[0].Subject != "second" || commits[1].Subject != "first" {
		t.Fatalf("ordering wrong: %#v", commits)
	}

	if commits[0].Author != "Reviewer" || commits[0].Email != "review@example.com" {
		t.Fatalf("author metadata missing: %#v", commits[0])
	}

	if commits[0].SHA == "" || commits[0].ShortSHA == "" || commits[0].Date == "" {
		t.Fatalf("sha/short/date missing: %#v", commits[0])
	}
}

func TestCheckoutBranchDirtyWorkingTree(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "--initial-branch=main")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")
	git(t, root, "config", "commit.gpgsign", "false")

	writeFile(t, root, "app.txt", "old\n")
	git(t, root, "add", "app.txt")
	git(t, root, "commit", "-m", "initial")

	git(t, root, "checkout", "-b", "other")
	git(t, root, "checkout", "main")

	writeFile(t, root, "app.txt", "old\nlocal change\n")

	err := CheckoutBranch(context.Background(), root, "other")

	if err == nil {
		t.Fatal("expected dirty working tree error, got nil")
	}

	var dirty *WorkingTreeDirtyError

	if !errors.As(err, &dirty) {
		t.Fatalf("error type = %T (%v), want *WorkingTreeDirtyError", err, err)
	}

	if len(dirty.Files) != 1 || dirty.Files[0] != "app.txt" {
		t.Fatalf("dirty files = %#v, want [app.txt]", dirty.Files)
	}

	current := strings.TrimSpace(gitOut(t, root, "branch", "--show-current"))

	if current != "main" {
		t.Fatalf("branch after failed checkout = %q, want main (working tree must be untouched)", current)
	}
}

func TestCheckoutBranchCleanWorkingTree(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "--initial-branch=main")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")
	git(t, root, "config", "commit.gpgsign", "false")

	writeFile(t, root, "app.txt", "hello\n")
	git(t, root, "add", "app.txt")
	git(t, root, "commit", "-m", "initial")

	git(t, root, "checkout", "-b", "feature")
	git(t, root, "checkout", "main")

	if err := CheckoutBranch(context.Background(), root, "feature"); err != nil {
		t.Fatalf("CheckoutBranch on clean tree failed: %v", err)
	}

	current := strings.TrimSpace(gitOut(t, root, "branch", "--show-current"))

	if current != "feature" {
		t.Fatalf("branch = %q, want feature", current)
	}
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()

	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}

	return string(out)
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func writeFile(t *testing.T, root string, path string, content string) {
	t.Helper()
	full := filepath.Join(root, path)

	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
