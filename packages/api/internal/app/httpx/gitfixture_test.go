package httpx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRepo is a real on-disk git repository created in t.TempDir() that
// tests can point handlers at. Use newGitRepo for an initialized repo with
// one initial commit on `main`.
type gitRepo struct {
	t    *testing.T
	Root string
}

// newGitRepo initialises a repo at t.TempDir() and seeds an initial commit
// on `main` so `git log` returns at least one entry. All operations use
// `--initial-branch=main` to make tests stable across git versions whose
// default branch differs.
func newGitRepo(t *testing.T) *gitRepo {
	t.Helper()

	root := t.TempDir()
	r := &gitRepo{t: t, Root: root}

	r.run("init", "--initial-branch=main", root)
	r.run("-C", root, "config", "user.email", "test@example.com")
	r.run("-C", root, "config", "user.name", "Test")
	r.run("-C", root, "config", "commit.gpgsign", "false")
	r.WriteFile("README.md", "# test repo\n")
	r.run("-C", root, "add", "README.md")
	r.run("-C", root, "commit", "-m", "initial")

	return r
}

// WriteFile writes contents to relPath inside the repo, creating parent dirs
// as needed.
func (r *gitRepo) WriteFile(relPath, contents string) {
	r.t.Helper()

	abs := filepath.Join(r.Root, relPath)

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		r.t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}

	if err := os.WriteFile(abs, []byte(contents), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", abs, err)
	}
}

// Commit stages the supplied paths (or all changes when empty) and creates
// a commit with the given message.
func (r *gitRepo) Commit(message string, paths ...string) string {
	r.t.Helper()

	if len(paths) == 0 {
		r.run("-C", r.Root, "add", "-A")
	} else {
		args := append([]string{"-C", r.Root, "add"}, paths...)
		r.run(args...)
	}

	r.run("-C", r.Root, "commit", "-m", message)

	return r.HeadSHA()
}

// Branch creates branch name pointing at HEAD without checking it out.
func (r *gitRepo) Branch(name string) {
	r.t.Helper()
	r.run("-C", r.Root, "branch", name)
}

// Checkout switches to branch name, creating it when create is true.
func (r *gitRepo) Checkout(name string, create bool) {
	r.t.Helper()

	if create {
		r.run("-C", r.Root, "checkout", "-b", name)

		return
	}

	r.run("-C", r.Root, "checkout", name)
}

// HeadSHA returns the current HEAD commit SHA.
func (r *gitRepo) HeadSHA() string {
	r.t.Helper()

	out, err := exec.Command("git", "-C", r.Root, "rev-parse", "HEAD").Output()

	if err != nil {
		r.t.Fatalf("rev-parse HEAD: %v", err)
	}

	return strings.TrimSpace(string(out))
}

func (r *gitRepo) run(args ...string) {
	r.t.Helper()

	cmd := exec.Command("git", args...)

	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
