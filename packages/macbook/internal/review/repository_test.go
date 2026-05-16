package review

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
