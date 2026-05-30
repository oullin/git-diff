package review

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRepositoryStateOnCleanRepo(t *testing.T) {
	repo := newGitRepo(t)

	state, err := ReadRepositoryState(context.Background(), repo)

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if state.Root != repo {
		t.Fatalf("expected root=%q, got %q", repo, state.Root)
	}

	if state.Mode != RepositoryModeWorking {
		t.Fatalf("expected mode=working, got %q", state.Mode)
	}

	if state.Branch != "main" {
		t.Fatalf("expected branch=main, got %q", state.Branch)
	}

	if len(state.Files) != 0 {
		t.Fatalf("expected zero changed files on clean repo, got %v", state.Files)
	}
}

func TestReadRepositoryStateDetectsUntrackedAndStaged(t *testing.T) {
	repo := newGitRepo(t)

	// Untracked file.
	if err := os.WriteFile(filepath.Join(repo, "untracked.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Staged file.
	if err := os.WriteFile(filepath.Join(repo, "staged.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	runGit(t, "-C", repo, "add", "staged.go")

	state, err := ReadRepositoryState(context.Background(), repo)

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	paths := map[string]GitFileStatus{}

	for _, f := range state.Files {
		paths[f.Path] = f.Status
	}

	if _, ok := paths["untracked.go"]; !ok {
		t.Fatalf("expected untracked.go in files: %v", paths)
	}

	if _, ok := paths["staged.go"]; !ok {
		t.Fatalf("expected staged.go in files: %v", paths)
	}
}

func TestReadRepositoryStateErrorsOnNonGitPath(t *testing.T) {
	tmp := t.TempDir()

	if _, err := ReadRepositoryState(context.Background(), tmp); err == nil {
		t.Fatalf("expected error for non-git path")
	}
}
