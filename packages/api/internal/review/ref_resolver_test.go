package review

import (
	"context"
	"strings"
	"testing"
)

func TestResolveCommitRefAcceptsShortAndFullSha(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	full := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD"))
	short := full[:7]

	cases := map[string]string{
		"full sha":  full,
		"short sha": short,
		"HEAD":      "HEAD",
	}

	for name, ref := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ResolveCommitRef(context.Background(), root, ref)

			if err != nil {
				t.Fatalf("resolve %q: %v", ref, err)
			}

			if got != full {
				t.Fatalf("resolved = %q, want %q", got, full)
			}
		})
	}
}

func TestResolveCommitRefHandlesHeadRevisionSyntax(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	// Build a second commit so HEAD~1 is reachable.
	writeFile(t, root, "second.txt", "hello\n")
	git(t, root, "add", "second.txt")
	git(t, root, "commit", "-m", "second")

	firstFull := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD~1"))
	got, err := ResolveCommitRef(context.Background(), root, "HEAD~1")

	if err != nil {
		t.Fatalf("HEAD~1: %v", err)
	}

	if got != firstFull {
		t.Fatalf("resolved HEAD~1 = %q, want %q", got, firstFull)
	}
}

func TestResolveCommitRefRejectsEmpty(t *testing.T) {
	if _, err := ResolveCommitRef(context.Background(), t.TempDir(), "   "); err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestResolveCommitRefSurfacesInvalidRef(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	_, err := ResolveCommitRef(context.Background(), root, "definitely-not-a-ref")

	if err == nil {
		t.Fatal("expected error for invalid ref")
	}

	if !strings.Contains(err.Error(), "definitely-not-a-ref") {
		t.Fatalf("error %q should mention the ref", err.Error())
	}
}

func TestResolveParentRefReturnsParentSha(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	firstFull := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD"))

	writeFile(t, root, "second.txt", "hello\n")
	git(t, root, "add", "second.txt")
	git(t, root, "commit", "-m", "second")

	parent, err := ResolveParentRef(context.Background(), root, "HEAD")

	if err != nil {
		t.Fatalf("parent: %v", err)
	}

	if parent != firstFull {
		t.Fatalf("parent = %q, want %q (first commit)", parent, firstFull)
	}
}

func TestResolveParentRefEmptyForRootCommit(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	parent, err := ResolveParentRef(context.Background(), root, "HEAD")

	if err != nil {
		t.Fatalf("parent: %v", err)
	}

	if parent != "" {
		t.Fatalf("root commit should have no parent, got %q", parent)
	}
}

func seedRepo(t *testing.T, root string) {
	t.Helper()

	git(t, root, "init")
	git(t, root, "config", "user.email", "ref@example.com")
	git(t, root, "config", "user.name", "Ref Tester")
	git(t, root, "config", "commit.gpgsign", "false")
	writeFile(t, root, "seed.txt", "seed\n")
	git(t, root, "add", "seed.txt")
	git(t, root, "commit", "-m", "seed")
}
