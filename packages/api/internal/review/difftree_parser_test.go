package review

import (
	"testing"
)

func TestDiffTreeParserSimpleAddDeleteModify(t *testing.T) {
	raw := []byte("A\x00added.go\x00D\x00removed.go\x00M\x00modified.go\x00")

	entries := DiffTreeParser{}.Parse(raw)

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d (%v)", len(entries), entries)
	}

	if entries[0].status != StatusAdded || entries[0].path != "added.go" {
		t.Fatalf("entry[0]: %+v", entries[0])
	}

	if entries[1].status != StatusDeleted || entries[1].path != "removed.go" {
		t.Fatalf("entry[1]: %+v", entries[1])
	}

	if entries[2].status != StatusModified || entries[2].path != "modified.go" {
		t.Fatalf("entry[2]: %+v", entries[2])
	}
}

func TestDiffTreeParserRenameTakesTwoExtraFields(t *testing.T) {
	raw := []byte("R\x00old.go\x00new.go\x00")

	entries := DiffTreeParser{}.Parse(raw)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d (%v)", len(entries), entries)
	}

	if entries[0].status != StatusRenamed {
		t.Fatalf("expected renamed, got %q", entries[0].status)
	}

	if entries[0].old != "old.go" || entries[0].path != "new.go" {
		t.Fatalf("unexpected paths: %+v", entries[0])
	}
}

func TestDiffTreeParserCopyTreatedAsRename(t *testing.T) {
	raw := []byte("C\x00src.go\x00dst.go\x00")

	entries := DiffTreeParser{}.Parse(raw)

	if len(entries) != 1 || entries[0].status != StatusRenamed {
		t.Fatalf("expected single renamed entry, got %v", entries)
	}
}

func TestDiffTreeParserUnknownStatusFallsBackToModified(t *testing.T) {
	raw := []byte("X\x00unknown.go\x00")

	entries := DiffTreeParser{}.Parse(raw)

	if len(entries) != 1 || entries[0].status != StatusModified {
		t.Fatalf("expected fallback to modified, got %v", entries)
	}
}

func TestDiffTreeParserSkipsEmptyAndTruncated(t *testing.T) {
	raw := []byte("\x00A\x00good.go\x00R\x00trailing") // trailing R has no completion

	entries := DiffTreeParser{}.Parse(raw)

	if len(entries) != 1 || entries[0].path != "good.go" {
		t.Fatalf("expected single 'good.go', got %v", entries)
	}
}
