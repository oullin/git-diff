package lineparse

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// SplitUnifiedPatch breaks a multi-file git patch into per-file sections,
// keyed by the b-side path (the "new" path, which differs from a-side for
// renames). Order of insertion in the returned map is not preserved — use
// SplitUnifiedPatchOrdered when call sites need a stable iteration.
//
// Each returned value is the raw bytes of the section, starting at its
// `diff --git ...` header and ending right before the next one. Trailing
// newlines from the source are preserved so the section round-trips back
// into git tooling unchanged.
//
// Empty input returns an empty (non-nil) map.

// PatchSection is one file's slice of a multi-file unified patch.
type PatchSection struct {
	// Path is the b-side path (the new path; same as a-side for non-renames).
	Path string
	// OldPath is the a-side path; equals Path when the file wasn't renamed.
	OldPath string
	// Body is the raw bytes of the section, starting at `diff --git`.
	Body []byte
}

func SplitUnifiedPatch(raw []byte) map[string][]byte {
	sections := SplitUnifiedPatchOrdered(raw)
	out := make(map[string][]byte, len(sections))

	for _, sec := range sections {
		out[sec.Path] = sec.Body
	}

	return out
}

// SplitUnifiedPatchOrdered returns one PatchSection per file in source order.
// Use this when callers need deterministic iteration (e.g. to preserve
// `git diff` ordering in the UI).
func SplitUnifiedPatchOrdered(raw []byte) []PatchSection {
	if len(raw) == 0 {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var sections []PatchSection

	var current *PatchSection

	var renameTo, renameFrom string

	flush := func() {
		if current == nil {
			return
		}

		if renameTo != "" {
			current.Path = renameTo
		}

		if renameFrom != "" {
			current.OldPath = renameFrom
		}

		if current.OldPath == "" {
			current.OldPath = current.Path
		}

		sections = append(sections, *current)
		current = nil
		renameTo = ""
		renameFrom = ""
	}

	for scanner.Scan() {
		line := scanner.Bytes()

		if bytes.HasPrefix(line, []byte("diff --git ")) {
			flush()

			a, b := parseDiffGitHeader(string(line))

			lineCopy := make([]byte, 0, len(line)+1)
			lineCopy = append(lineCopy, line...)
			lineCopy = append(lineCopy, '\n')
			current = &PatchSection{Path: b, OldPath: a, Body: lineCopy}

			continue
		}

		if current == nil {
			// Skip preamble noise before the first `diff --git`.
			continue
		}

		if bytes.HasPrefix(line, []byte("rename to ")) {
			renameTo = string(bytes.TrimPrefix(line, []byte("rename to ")))
		}

		if bytes.HasPrefix(line, []byte("rename from ")) {
			renameFrom = string(bytes.TrimPrefix(line, []byte("rename from ")))
		}

		current.Body = append(current.Body, line...)
		current.Body = append(current.Body, '\n')
	}

	flush()

	// Strip the trailing newline we always append if the source didn't end with one.
	// scanner.Scan strips the line terminator, so we re-add `\n` per line; if the
	// input ended without a final `\n`, our last line gets an extra one. Detect
	// that by checking the source's final byte and trimming the section body.
	if len(sections) > 0 && raw[len(raw)-1] != '\n' {
		last := &sections[len(sections)-1]

		if n := len(last.Body); n > 0 && last.Body[n-1] == '\n' {
			last.Body = last.Body[:n-1]
		}
	}

	return sections
}

// parseDiffGitHeader extracts the a/ and b/ paths from a `diff --git` header
// line. Handles both unquoted (`a/path`) and quoted (`"a/path with space"`)
// forms; in the quoted form, basic C-style escapes are decoded.
//
// Returns ("", "") when the line doesn't parse — callers should fall back
// to the `rename to` / `--- a/` / `+++ b/` lines that follow.
func parseDiffGitHeader(line string) (oldPath, newPath string) {
	// Strip the `diff --git ` prefix.
	const prefix = "diff --git "

	if !strings.HasPrefix(line, prefix) {
		return "", ""
	}

	rest := line[len(prefix):]

	// Walk left-to-right consuming the a/ token, then the b/ token.
	a, after := consumePath(rest)

	if a == "" {
		return "", ""
	}

	after = strings.TrimLeft(after, " ")
	b, _ := consumePath(after)

	if b == "" {
		return "", ""
	}

	return stripABPrefix(a, "a/"), stripABPrefix(b, "b/")
}

// consumePath returns the next path token (quoted or unquoted) and the
// remainder of the line.
func consumePath(s string) (path, rest string) {
	if s == "" {
		return "", ""
	}

	if s[0] == '"' {
		// Find the matching unescaped closing quote.
		end := -1

		for i := 1; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				i++

				continue
			}

			if s[i] == '"' {
				end = i

				break
			}
		}

		if end < 0 {
			return "", ""
		}

		decoded, err := strconv.Unquote(s[:end+1])

		if err != nil {
			// Fall back to the raw quoted body.
			decoded = s[1:end]
		}

		return decoded, s[end+1:]
	}

	// Unquoted: read up to next space.
	if idx := strings.IndexByte(s, ' '); idx >= 0 {
		return s[:idx], s[idx:]
	}

	return s, ""
}

func stripABPrefix(p, prefix string) string {
	if strings.HasPrefix(p, prefix) {
		return p[len(prefix):]
	}

	return p
}
