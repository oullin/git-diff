package review

import (
	"bytes"
)

// DiffTreeParser turns the raw bytes of
// `git diff-tree --name-status -z` into a slice of commitDiffEntry.
type DiffTreeParser struct{}

func (DiffTreeParser) Parse(raw []byte) []commitDiffEntry {
	parts := bytes.Split(raw, []byte{0})
	entries := []commitDiffEntry{}

	for i := 0; i < len(parts); i++ {
		field := string(parts[i])

		if field == "" {
			continue
		}

		status := field[0]
		entry := commitDiffEntry{}

		switch status {
		case 'A':
			entry.status = StatusAdded
		case 'D':
			entry.status = StatusDeleted
		case 'R', 'C':
			entry.status = StatusRenamed
		default:
			entry.status = StatusModified
		}

		if status == 'R' || status == 'C' {
			if i+2 >= len(parts) {
				return entries
			}

			entry.old = string(parts[i+1])
			entry.path = string(parts[i+2])
			i += 2
		} else {
			if i+1 >= len(parts) {
				return entries
			}

			entry.path = string(parts[i+1])
			i++
		}

		if entry.path == "" {
			continue
		}

		entries = append(entries, entry)
	}

	return entries
}

func parseDiffTreeNameStatus(raw []byte) []commitDiffEntry {
	return DiffTreeParser{}.Parse(raw)
}
