package review

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newGitRepo creates a one-commit git repo at t.TempDir() on main. The
// returned path is the resolved top-level path (matches what `RootFor`
// reports), avoiding `/private` macOS symlink mismatches.
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

	return resolveRepoTopLevel(t, root)
}

func resolveRepoTopLevel(t *testing.T, path string) string {
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
