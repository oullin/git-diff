package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newGitRepo creates an on-disk git repository with one initial commit on
// `main`. The service layer's BranchService calls into `internal/review`
// which shells out to `git`, so we need a real repo on disk.
//
// The returned path is the *resolved* top-level path that `git rev-parse
// --show-toplevel` reports — on macOS this differs from `t.TempDir()` by a
// `/private` prefix, and BranchService uses the resolved form when looking
// up the repository row.
func newGitRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	runGit(t, "init", "--initial-branch=main", root)
	runGit(t, "-C", root, "config", "user.email", "test@example.com")
	runGit(t, "-C", root, "config", "user.name", "Test")
	runGit(t, "-C", root, "config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	runGit(t, "-C", root, "add", "README.md")
	runGit(t, "-C", root, "commit", "-m", "initial")

	return resolveTopLevel(t, root)
}

func resolveTopLevel(t *testing.T, path string) string {
	t.Helper()

	out, err := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel").Output()

	if err != nil {
		t.Fatalf("resolve top-level for %q: %v", path, err)
	}

	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
