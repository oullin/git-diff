package walks

import (
	"strings"
	"testing"

	"github.com/oullin/git-diff/internal/review"
)

func TestEnforceBudgetTruncatesOversizedFile(t *testing.T) {
	files := []review.ChangedFile{
		{
			Path: "big.go",
			Sections: []review.DiffSection{
				{Kind: "patch", Patch: strings.Repeat("a", 5000)},
			},
		},
	}

	out := enforceBudget(files, Budget{PerFileBytes: 100, TotalBytes: 10_000})

	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}

	if !strings.HasSuffix(out[0].PatchBody, truncationMarker) {
		t.Fatalf("expected truncation marker appended, got %q", out[0].PatchBody)
	}

	if len(out[0].PatchBody) > 100+len(truncationMarker)+1 {
		t.Fatalf("body exceeded PerFileBytes + marker: %d bytes", len(out[0].PatchBody))
	}
}

func TestEnforceBudgetStopsOnTotalAndSetsOmittedAfter(t *testing.T) {
	files := []review.ChangedFile{
		{Path: "a.go", Sections: []review.DiffSection{{Patch: strings.Repeat("a", 200)}}},
		{Path: "b.go", Sections: []review.DiffSection{{Patch: strings.Repeat("b", 200)}}},
		{Path: "c.go", Sections: []review.DiffSection{{Patch: strings.Repeat("c", 200)}}},
	}

	out := enforceBudget(files, Budget{PerFileBytes: 1024, TotalBytes: 300})

	if len(out) >= len(files) {
		t.Fatalf("expected truncation to drop tail, got %d entries", len(out))
	}

	last := out[len(out)-1]

	if !last.OmittedAfter {
		t.Fatalf("expected OmittedAfter on last entry, got %+v", last)
	}
}

func TestEnforceBudgetDefaultsZeroValues(t *testing.T) {
	files := []review.ChangedFile{{Path: "a.go", Sections: []review.DiffSection{{Patch: "hi"}}}}

	out := enforceBudget(files, Budget{})

	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}

	if out[0].PatchBody == "" {
		t.Fatalf("expected non-empty body")
	}
}

func TestConcatenatePatchesMarksBinarySection(t *testing.T) {
	file := review.ChangedFile{
		Sections: []review.DiffSection{
			{Kind: "binary", Binary: true},
			{Kind: "text", Patch: "+foo\n"},
		},
	}

	out := concatenatePatches(file)

	if !strings.Contains(out, "[binary section: binary]") {
		t.Fatalf("expected binary marker, got %q", out)
	}

	if !strings.Contains(out, "+foo") {
		t.Fatalf("expected text section preserved, got %q", out)
	}
}
