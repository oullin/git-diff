package review

import (
	"bytes"
)

// StatusParser turns the raw bytes of `git status --porcelain=v1 -z` into
// a slice of statusEntry.
type StatusParser struct{}

func (StatusParser) Parse(raw []byte) []statusEntry {
	parts := bytes.Split(raw, []byte{0})
	entries := []statusEntry{}

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if len(part) < 4 {
			continue
		}

		entry := statusEntry{index: part[0], work: part[1], path: string(part[3:])}

		if entry.index == 'R' || entry.index == 'C' {
			entry.rename = true

			if i+1 < len(parts) {
				i++
				entry.old = entry.path
				entry.path = string(parts[i])
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

func parseStatus(raw []byte) []statusEntry {
	return StatusParser{}.Parse(raw)
}
