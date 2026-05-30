package review

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListBranchesReturnsMain(t *testing.T) {
	repo := newGitRepo(t)

	branches, err := ListBranches(context.Background(), repo)

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	found := false

	for _, name := range branches {
		if name == "main" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected main in branches, got %v", branches)
	}
}

func TestCreateBranchSwitchesToIt(t *testing.T) {
	repo := newGitRepo(t)

	if err := CreateBranch(context.Background(), repo, "feature/x"); err != nil {
		t.Fatalf("create: %v", err)
	}

	branches, _ := ListBranches(context.Background(), repo)

	found := false

	for _, name := range branches {
		if name == "feature/x" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected feature/x in branches, got %v", branches)
	}
}

func TestCreateBranchRejectsInvalidName(t *testing.T) {
	repo := newGitRepo(t)

	if err := CreateBranch(context.Background(), repo, ""); err == nil {
		t.Fatalf("expected error for empty name")
	}

	if err := CreateBranch(context.Background(), repo, "-bad"); err == nil {
		t.Fatalf("expected error for name starting with '-'")
	}

	if err := CreateBranch(context.Background(), repo, "has spaces"); err == nil {
		t.Fatalf("expected error for whitespace in name")
	}

	if err := CreateBranch(context.Background(), repo, "has..dots"); err == nil {
		t.Fatalf("expected error for '..' in name")
	}
}

func TestDeleteBranchRefusesCurrentBranch(t *testing.T) {
	repo := newGitRepo(t)

	if err := CreateBranch(context.Background(), repo, "feature/y"); err != nil {
		t.Fatalf("create: %v", err)
	}

	// CreateBranch checks the branch out, so deleting feature/y must fail.
	err := DeleteBranch(context.Background(), repo, "feature/y")

	if err == nil {
		t.Fatalf("expected error deleting current branch")
	}

	if !strings.Contains(err.Error(), "currently checked-out") {
		t.Fatalf("expected currently-checked-out error, got %v", err)
	}
}

func TestCheckoutBranchSucceeds(t *testing.T) {
	repo := newGitRepo(t)
	runGit(t, "-C", repo, "branch", "feature/z")

	if err := CheckoutBranch(context.Background(), repo, "feature/z"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
}

func TestCheckoutBranchDetectsDirtyTree(t *testing.T) {
	repo := newGitRepo(t)
	runGit(t, "-C", repo, "branch", "feature/w")

	// Make the tree dirty by modifying a tracked file.
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# dirty\n"), 0o644); err != nil {
		t.Fatalf("dirty: %v", err)
	}

	err := CheckoutBranch(context.Background(), repo, "feature/w")

	var dirty *WorkingTreeDirtyError

	if !errors.As(err, &dirty) {
		t.Fatalf("expected WorkingTreeDirtyError, got %v", err)
	}

	if len(dirty.Files) == 0 {
		t.Fatalf("expected at least one dirty file listed")
	}
}
