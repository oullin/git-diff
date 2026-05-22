package walkthrough

import (
	"fmt"
	"strings"

	"github.com/gocanto/git-diff/internal/review"
)

// truncatedFile is the post-budget shape the prompt builder consumes.
// Keeps the prompt builder oblivious to budget details — it just renders
// whatever this function hands back.
type truncatedFile struct {
	File         review.ChangedFile
	PatchBody    string
	OmittedAfter bool // true on the last file when the total budget cut the rest
}

// enforceBudget walks the state's files in order, truncating each file's
// concatenated patch to PerFileBytes and stopping the iteration once the
// cumulative size crosses TotalBytes. Pure function — no IO, easy to
// unit test against captured states.
//
// When a file is truncated, "[…truncated]" is appended so the LLM
// understands the body was cut. When the total budget terminates the
// iteration, OmittedAfter is set on the final included file so the
// caller can render a global marker.
func enforceBudget(files []review.ChangedFile, budget Budget) []truncatedFile {
	if budget.PerFileBytes <= 0 {
		budget.PerFileBytes = DefaultBudget().PerFileBytes
	}

	if budget.TotalBytes <= 0 {
		budget.TotalBytes = DefaultBudget().TotalBytes
	}

	out := make([]truncatedFile, 0, len(files))
	total := 0

	for i, file := range files {
		body := concatenatePatches(file)

		if len(body) > budget.PerFileBytes {
			body = body[:budget.PerFileBytes] + truncationMarker
		}

		total += len(body)
		entry := truncatedFile{File: file, PatchBody: body}

		if total > budget.TotalBytes {
			entry.OmittedAfter = i+1 < len(files)
			out = append(out, entry)

			return out
		}

		out = append(out, entry)
	}

	return out
}

// concatenatePatches joins every diff section in the file, marking
// binary sections so the LLM doesn't try to interpret the bytes.
func concatenatePatches(file review.ChangedFile) string {
	var b strings.Builder

	for _, section := range file.Sections {
		if section.Binary {
			fmt.Fprintf(&b, "[binary section: %s]\n", section.Kind)

			continue
		}

		b.WriteString(section.Patch)
		b.WriteString("\n")
	}

	return b.String()
}

const truncationMarker = "\n[…truncated]\n"
