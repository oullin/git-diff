package review

import (
	"context"
	"strings"
	"testing"
)

func TestReadRepositoryFileRange_WorkingTree(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")

	content := strings.Join([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
	}, "\n") + "\n"
	writeFile(t, root, "doc.txt", content)

	t.Run("returns the requested slice", func(t *testing.T) {
		got, err := ReadRepositoryFileRange(context.Background(), root, "doc.txt", "", 2, 4)

		if err != nil {
			t.Fatal(err)
		}

		want := []string{"line 2", "line 3", "line 4"}

		if len(got.Lines) != len(want) {
			t.Fatalf("len = %d, want %d", len(got.Lines), len(want))
		}

		for i, line := range want {
			if got.Lines[i] != line {
				t.Errorf("line[%d] = %q, want %q", i, got.Lines[i], line)
			}
		}

		if got.StartLine != 2 || got.EndLine != 4 {
			t.Errorf("start/end = %d/%d, want 2/4", got.StartLine, got.EndLine)
		}

		if got.EOF {
			t.Error("eof = true, want false (line 5 still unread)")
		}
	})

	t.Run("clamps endLine to file end and reports eof", func(t *testing.T) {
		got, err := ReadRepositoryFileRange(context.Background(), root, "doc.txt", "", 4, 99)

		if err != nil {
			t.Fatal(err)
		}

		if got.EndLine != 5 {
			t.Errorf("endLine = %d, want 5", got.EndLine)
		}

		if !got.EOF {
			t.Error("eof = false, want true")
		}
	})

	t.Run("returns empty slice past EOF", func(t *testing.T) {
		got, err := ReadRepositoryFileRange(context.Background(), root, "doc.txt", "", 99, 100)

		if err != nil {
			t.Fatal(err)
		}

		if len(got.Lines) != 0 {
			t.Errorf("lines = %d, want 0", len(got.Lines))
		}

		if !got.EOF {
			t.Error("eof = false, want true")
		}
	})

	t.Run("rejects bogus ranges", func(t *testing.T) {
		_, err := ReadRepositoryFileRange(context.Background(), root, "doc.txt", "", 5, 2)

		if err == nil {
			t.Fatal("expected error for endLine < startLine")
		}
	})
}

func TestReadRepositoryFileRange_CommitRef(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "review@example.com")
	git(t, root, "config", "user.name", "Reviewer")

	writeFile(t, root, "doc.txt", "original 1\noriginal 2\noriginal 3\n")
	git(t, root, "add", "doc.txt")
	git(t, root, "commit", "-m", "initial")

	commitSHA := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD"))

	// Overwrite working copy so the ref-based read is provably different.
	writeFile(t, root, "doc.txt", "different\n")

	got, err := ReadRepositoryFileRange(context.Background(), root, "doc.txt", commitSHA, 1, 3)

	if err != nil {
		t.Fatal(err)
	}

	want := []string{"original 1", "original 2", "original 3"}

	if len(got.Lines) != len(want) {
		t.Fatalf("len = %d, want %d", len(got.Lines), len(want))
	}

	for i, line := range want {
		if got.Lines[i] != line {
			t.Errorf("line[%d] = %q, want %q", i, got.Lines[i], line)
		}
	}
}
