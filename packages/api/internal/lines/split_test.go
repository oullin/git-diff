package lines

import (
	"bytes"
	"testing"
)

func TestSplitUnifiedPatchSimpleModification(t *testing.T) {
	raw := []byte(`diff --git a/file.txt b/file.txt
index 0123abc..def4567 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,4 @@
 line1
-line2
+line2-modified
+line3
 line4
`)

	got := SplitUnifiedPatch(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}

	body, ok := got["file.txt"]

	if !ok {
		t.Fatalf("missing file.txt key; got keys %v", keys(got))
	}

	if !bytes.Equal(body, raw) {
		t.Fatalf("single-file body should round-trip; got %q", body)
	}
}

func TestSplitUnifiedPatchMultipleFiles(t *testing.T) {
	raw := []byte(`diff --git a/first.txt b/first.txt
index a..b 100644
--- a/first.txt
+++ b/first.txt
@@ -1,1 +1,1 @@
-old
+new
diff --git a/second.txt b/second.txt
index c..d 100644
--- a/second.txt
+++ b/second.txt
@@ -1,1 +1,2 @@
 keep
+added
`)

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(got))
	}

	if got[0].Path != "first.txt" {
		t.Fatalf("first.Path = %q", got[0].Path)
	}

	if got[1].Path != "second.txt" {
		t.Fatalf("second.Path = %q", got[1].Path)
	}

	if !bytes.HasPrefix(got[0].Body, []byte("diff --git a/first.txt")) {
		t.Fatalf("first section should start with its header; got %q", string(got[0].Body[:40]))
	}

	if bytes.Contains(got[0].Body, []byte("diff --git a/second")) {
		t.Fatalf("first section leaked into second; got %q", got[0].Body)
	}
}

func TestSplitUnifiedPatchRename(t *testing.T) {
	raw := []byte(`diff --git a/old/path.txt b/new/path.txt
similarity index 95%
rename from old/path.txt
rename to new/path.txt
index abc..def 100644
--- a/old/path.txt
+++ b/new/path.txt
@@ -1,2 +1,2 @@
-foo
+bar
 baz
`)

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	if got[0].Path != "new/path.txt" {
		t.Fatalf("Path should be the new b/ path, got %q", got[0].Path)
	}

	if got[0].OldPath != "old/path.txt" {
		t.Fatalf("OldPath should be the old a/ path, got %q", got[0].OldPath)
	}
}

func TestSplitUnifiedPatchPureRenameNoHunks(t *testing.T) {
	raw := []byte(`diff --git a/old.txt b/new.txt
similarity index 100%
rename from old.txt
rename to new.txt
`)

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	if got[0].Path != "new.txt" || got[0].OldPath != "old.txt" {
		t.Fatalf("rename mismatched: %+v", got[0])
	}
}

func TestSplitUnifiedPatchDeletion(t *testing.T) {
	raw := []byte(`diff --git a/gone.txt b/gone.txt
deleted file mode 100644
index abc..0000000
--- a/gone.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-line1
-line2
`)

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	if got[0].Path != "gone.txt" {
		t.Fatalf("deletion Path = %q", got[0].Path)
	}
}

func TestSplitUnifiedPatchBinaryFile(t *testing.T) {
	raw := []byte(`diff --git a/img.png b/img.png
index abc..def 100644
GIT binary patch
delta 12
zcmZ...

diff --git a/next.txt b/next.txt
index 1..2 100644
--- a/next.txt
+++ b/next.txt
@@ -1 +1 @@
-old
+new
`)

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 2 {
		t.Fatalf("expected 2 sections (binary + text), got %d", len(got))
	}

	if got[0].Path != "img.png" {
		t.Fatalf("binary section path = %q", got[0].Path)
	}

	if !bytes.Contains(got[0].Body, []byte("GIT binary patch")) {
		t.Fatalf("binary body should retain its marker; got %q", got[0].Body)
	}
}

func TestSplitUnifiedPatchQuotedPaths(t *testing.T) {
	// Git emits quoted paths when the path contains characters that need
	// escaping (spaces, non-ASCII, etc.).
	raw := []byte("diff --git \"a/path with space.txt\" \"b/path with space.txt\"\n" +
		"index a..b 100644\n" +
		"--- \"a/path with space.txt\"\n" +
		"+++ \"b/path with space.txt\"\n" +
		"@@ -1 +1 @@\n" +
		"-old\n" +
		"+new\n")

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	if got[0].Path != "path with space.txt" {
		t.Fatalf("quoted path = %q", got[0].Path)
	}
}

func TestSplitUnifiedPatchEmptyInput(t *testing.T) {
	if got := SplitUnifiedPatch(nil); len(got) != 0 {
		t.Fatalf("nil input should produce empty map, got %v", got)
	}

	if got := SplitUnifiedPatch([]byte("")); len(got) != 0 {
		t.Fatalf("empty input should produce empty map, got %v", got)
	}
}

func TestSplitUnifiedPatchPreservesByteFidelityWithoutTrailingNewline(t *testing.T) {
	raw := []byte("diff --git a/x b/x\n+++ b/x\n@@ -0,0 +1 @@\n+only")

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	if !bytes.Equal(got[0].Body, raw) {
		t.Fatalf("body should equal input; got %q want %q", got[0].Body, raw)
	}
}

func TestSplitUnifiedPatchHandlesRenameWithQuotedPaths(t *testing.T) {
	raw := []byte("diff --git \"a/old name.txt\" \"b/new name.txt\"\n" +
		"similarity index 100%\n" +
		"rename from \"old name.txt\"\n" +
		"rename to \"new name.txt\"\n")

	got := SplitUnifiedPatchOrdered(raw)

	if len(got) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got))
	}

	// `rename to` carries the path verbatim (with quotes if git emitted them);
	// the splitter trusts what git writes.
	if got[0].Path == "" {
		t.Fatalf("rename path should not be empty: %+v", got[0])
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))

	for k := range m {
		out = append(out, k)
	}

	return out
}
