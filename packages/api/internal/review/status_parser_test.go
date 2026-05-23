package review

import (
	"testing"
)

func TestStatusParserHandlesPlainModifications(t *testing.T) {
	// `git status --porcelain=v1 -z` emits one NUL-terminated record per
	// path. The first two bytes are the index/work flags and a space
	// separator is the third byte.
	raw := []byte(" M src/a.go\x00?? new.txt\x00")
	got := StatusParser{}.Parse(raw)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d (%#v)", len(got), got)
	}

	if got[0].index != ' ' || got[0].work != 'M' || got[0].path != "src/a.go" {
		t.Fatalf("first entry mismatched: %#v", got[0])
	}

	if got[1].index != '?' || got[1].work != '?' || got[1].path != "new.txt" {
		t.Fatalf("untracked entry mismatched: %#v", got[1])
	}
}

func TestStatusParserHandlesRename(t *testing.T) {
	// Renames are emitted as two NUL-terminated fields: the index/work +
	// destination path, then the source path.
	raw := []byte("R  new/path.go\x00old/path.go\x00")
	got := StatusParser{}.Parse(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d (%#v)", len(got), got)
	}

	entry := got[0]

	if !entry.rename {
		t.Fatalf("expected rename flag")
	}

	if entry.path != "old/path.go" {
		t.Fatalf("expected path = old/path.go, got %q", entry.path)
	}

	if entry.old != "new/path.go" {
		t.Fatalf("expected old = new/path.go, got %q", entry.old)
	}
}

// Regression: codiff v0.7.0 fixed a bug where stashed changes displayed as
// unstaged modifications. Our backend reads working-tree state exclusively via
// `git status --porcelain=v1 -z` (see state_reader.go), which by design omits
// stashed entries — so a clean tree after `git stash` produces empty output.
// This test locks in that the parser turns that empty input into zero entries,
// keeping any future "let's enrich status output" change from regressing.
func TestStatusParserReturnsNoEntriesForCleanTree(t *testing.T) {
	got := StatusParser{}.Parse(nil)

	if len(got) != 0 {
		t.Fatalf("expected 0 entries for empty input, got %d (%#v)", len(got), got)
	}

	got = StatusParser{}.Parse([]byte{})

	if len(got) != 0 {
		t.Fatalf("expected 0 entries for zero-length input, got %d (%#v)", len(got), got)
	}
}

func TestDiffTreeParserHandlesEachStatus(t *testing.T) {
	// `git diff-tree --name-status -z` emits flag\0path\0 for A/D/M, and
	// flag\0oldPath\0newPath\0 for R/C.
	raw := []byte("A\x00added.go\x00D\x00gone.go\x00R100\x00old.go\x00new.go\x00M\x00mod.go\x00")
	got := DiffTreeParser{}.Parse(raw)

	if len(got) != 4 {
		t.Fatalf("expected 4 entries, got %d (%#v)", len(got), got)
	}

	want := []commitDiffEntry{
		{status: StatusAdded, path: "added.go"},
		{status: StatusDeleted, path: "gone.go"},
		{status: StatusRenamed, path: "new.go", old: "old.go"},
		{status: StatusModified, path: "mod.go"},
	}

	for i, entry := range got {
		if entry != want[i] {
			t.Fatalf("entry %d mismatched: got %#v, want %#v", i, entry, want[i])
		}
	}
}
