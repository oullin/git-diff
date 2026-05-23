package walks

import (
	"fmt"
	"strings"

	"github.com/oullin/git-diff/internal/review"
)

type truncatedFile struct {
	File         review.ChangedFile
	PatchBody    string
	OmittedAfter bool // set on the last file when the total budget cut the rest
}

// enforceBudget caps each file at PerFileBytes (appending "[…truncated]")
// and stops once the running total crosses TotalBytes (setting
// OmittedAfter on the final included file).
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

// concatenatePatches marks binary sections so the LLM doesn't try to
// interpret the bytes.
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
