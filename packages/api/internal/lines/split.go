package lines

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// PatchSection is one file's slice of a multi-file unified patch.
// Path is the b-side (new) path; OldPath is the a-side (equal for
// non-renames). Body starts at `diff --git` and preserves trailing
// newlines so the section round-trips through git tooling unchanged.
type PatchSection struct {
	Path    string
	OldPath string
	Body    []byte
}

// SplitUnifiedPatch keys by b-side path; iteration order is undefined.
// Use SplitUnifiedPatchOrdered when call sites need stable iteration.
func SplitUnifiedPatch(raw []byte) map[string][]byte {
	sections := SplitUnifiedPatchOrdered(raw)
	out := make(map[string][]byte, len(sections))

	for _, sec := range sections {
		out[sec.Path] = sec.Body
	}

	return out
}

// SplitUnifiedPatchOrdered returns sections in source order so callers
// can preserve `git diff` ordering.
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

	// scanner.Scan strips line terminators and we re-add `\n` per line,
	// so an input without a trailing newline picks up an extra one — trim it.
	if len(sections) > 0 && raw[len(raw)-1] != '\n' {
		last := &sections[len(sections)-1]

		if n := len(last.Body); n > 0 && last.Body[n-1] == '\n' {
			last.Body = last.Body[:n-1]
		}
	}

	return sections
}

// parseDiffGitHeader handles both unquoted (`a/path`) and quoted
// (`"a/path with space"`) forms; quoted form decodes C-style escapes.
// Returns ("", "") on parse failure — callers should fall back to the
// `rename to` / `--- a/` / `+++ b/` lines that follow.
func parseDiffGitHeader(line string) (oldPath, newPath string) {
	const prefix = "diff --git "

	if !strings.HasPrefix(line, prefix) {
		return "", ""
	}

	rest := line[len(prefix):]

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

func consumePath(s string) (path, rest string) {
	if s == "" {
		return "", ""
	}

	if s[0] == '"' {
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
			decoded = s[1:end]
		}

		return decoded, s[end+1:]
	}

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
