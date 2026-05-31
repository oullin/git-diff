package review

import (
	"strings"
	"testing"
)

func TestCountPatchLines(t *testing.T) {
	patch := strings.Join([]string{
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,3 +1,4 @@",
		"-old line",
		"+new line",
		"+another",
		" context",
	}, "\n")

	added, deleted := countPatchLines(patch)

	if added != 2 {
		t.Fatalf("expected 2 additions, got %d", added)
	}

	if deleted != 1 {
		t.Fatalf("expected 1 deletion, got %d", deleted)
	}
}

func TestIsBinaryPatchDetectsBothMarkers(t *testing.T) {
	cases := map[string]bool{
		"":                              false,
		"context\n+foo\n":               false,
		"Binary files a and b differ":   true,
		"context\nGIT binary patch\n..": true,
	}

	for input, want := range cases {
		if got := isBinaryPatch(input); got != want {
			t.Fatalf("isBinaryPatch(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestFingerprintIsDeterministic(t *testing.T) {
	file := ChangedFile{
		Path:    "foo.go",
		OldPath: "old.go",
		Status:  StatusRenamed,
		Sections: []DiffSection{
			{Kind: "patch", Patch: "+hi\n"},
		},
	}

	a := fingerprint(file)
	b := fingerprint(file)

	if a != b {
		t.Fatalf("fingerprint not deterministic: %q vs %q", a, b)
	}

	if len(a) != 40 {
		t.Fatalf("expected 40-char sha1 hex, got %d (%q)", len(a), a)
	}
}

func TestFingerprintChangesWithContent(t *testing.T) {
	a := fingerprint(ChangedFile{Path: "foo.go", Sections: []DiffSection{{Patch: "+a"}}})
	b := fingerprint(ChangedFile{Path: "foo.go", Sections: []DiffSection{{Patch: "+b"}}})

	if a == b {
		t.Fatalf("expected different fingerprints for different patches")
	}
}

func TestPathSectionID(t *testing.T) {
	file := ChangedFile{Path: "src/main.go"}

	if got := pathSectionID(file, "staged"); got != "staged:src/main.go" {
		t.Fatalf("unexpected id %q", got)
	}
}
