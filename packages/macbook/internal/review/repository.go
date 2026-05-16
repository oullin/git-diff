package review

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type GitFileStatus string

type DiffSection struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Patch  string `json:"patch"`
	Binary bool   `json:"binary"`
}

type ChangedFile struct {
	Path        string        `json:"path"`
	OldPath     string        `json:"oldPath,omitempty"`
	Status      GitFileStatus `json:"status"`
	Additions   int           `json:"additions"`
	Deletions   int           `json:"deletions"`
	Binary      bool          `json:"binary"`
	Fingerprint string        `json:"fingerprint"`
	Sections    []DiffSection `json:"sections"`
}

type RepositoryState struct {
	Root         string        `json:"root"`
	LaunchPath   string        `json:"launchPath"`
	Branch       string        `json:"branch"`
	HeadSHA      string        `json:"headSha"`
	GeneratedAt  string        `json:"generatedAt"`
	Files        []ChangedFile `json:"files"`
	TrackedFiles []string      `json:"trackedFiles"`
	Additions    int           `json:"additions"`
	Deletions    int           `json:"deletions"`
}

type statusEntry struct {
	index  byte
	work   byte
	path   string
	old    string
	rename bool
}

const (
	StatusAdded     GitFileStatus = "added"
	StatusDeleted   GitFileStatus = "deleted"
	StatusModified  GitFileStatus = "modified"
	StatusRenamed   GitFileStatus = "renamed"
	StatusUntracked GitFileStatus = "untracked"
)

func ReadRepositoryState(ctx context.Context, launchPath string) (RepositoryState, error) {
	root, err := gitOutput(ctx, launchPath, "rev-parse", "--show-toplevel")

	if err != nil {
		return RepositoryState{}, fmt.Errorf("resolve git root: %w", err)
	}

	root = strings.TrimSpace(root)

	branch, _ := gitOutput(ctx, root, "branch", "--show-current")
	head, _ := gitOutput(ctx, root, "rev-parse", "--short", "HEAD")
	statusRaw, err := gitBytes(ctx, root, "status", "--porcelain=v1", "-z")

	if err != nil {
		return RepositoryState{}, fmt.Errorf("read git status: %w", err)
	}

	entries := parseStatus(statusRaw)
	files := make([]ChangedFile, 0, len(entries))

	for _, entry := range entries {
		file := ChangedFile{
			Path:    entry.path,
			OldPath: entry.old,
			Status:  statusFromEntry(entry),
		}

		if entry.index != ' ' && entry.index != '?' {
			patch, binary := filePatch(ctx, root, true, entry.path)
			file.Sections = append(file.Sections, DiffSection{ID: file.pathSectionID("staged"), Kind: "staged", Patch: patch, Binary: binary})
			file.Binary = file.Binary || binary
			add, del := countPatchLines(patch)
			file.Additions += add
			file.Deletions += del
		}

		if entry.work != ' ' && entry.work != '?' {
			patch, binary := filePatch(ctx, root, false, entry.path)
			file.Sections = append(file.Sections, DiffSection{ID: file.pathSectionID("unstaged"), Kind: "unstaged", Patch: patch, Binary: binary})
			file.Binary = file.Binary || binary
			add, del := countPatchLines(patch)
			file.Additions += add
			file.Deletions += del
		}

		if entry.index == '?' && entry.work == '?' {
			patch, binary := untrackedPatch(root, entry.path)
			file.Sections = append(file.Sections, DiffSection{ID: file.pathSectionID("untracked"), Kind: "untracked", Patch: patch, Binary: binary})
			file.Binary = binary
			file.Additions, file.Deletions = countPatchLines(patch)
		}

		if len(file.Sections) == 0 {
			continue
		}

		file.Fingerprint = fingerprint(file)
		files = append(files, file)
	}

	sort.SliceStable(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	tracked, _ := listTrackedFiles(ctx, root)

	state := RepositoryState{
		Root:         root,
		LaunchPath:   launchPath,
		Branch:       strings.TrimSpace(branch),
		HeadSHA:      strings.TrimSpace(head),
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339Nano),
		Files:        files,
		TrackedFiles: tracked,
	}

	for _, file := range files {
		state.Additions += file.Additions
		state.Deletions += file.Deletions
	}

	return state, nil
}

func parseStatus(raw []byte) []statusEntry {
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

func statusFromEntry(entry statusEntry) GitFileStatus {
	if entry.index == '?' && entry.work == '?' {
		return StatusUntracked
	}

	if entry.index == 'R' || entry.work == 'R' || entry.rename {
		return StatusRenamed
	}

	if entry.index == 'A' || entry.work == 'A' {
		return StatusAdded
	}

	if entry.index == 'D' || entry.work == 'D' {
		return StatusDeleted
	}

	return StatusModified
}

func filePatch(ctx context.Context, root string, staged bool, path string) (string, bool) {
	args := []string{"diff", "--binary", "--find-renames"}

	if staged {
		args = append(args, "--cached")
	}

	args = append(args, "--", path)
	patch, err := gitOutput(ctx, root, args...)

	if err != nil {
		return "", false
	}

	return patch, isBinaryPatch(patch)
}

func untrackedPatch(root string, path string) (string, bool) {
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	content, err := os.ReadFile(fullPath)

	if err != nil || !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
		return "", true
	}

	lines := strings.SplitAfter(string(content), "\n")

	var builder strings.Builder

	builder.WriteString("diff --git a/")
	builder.WriteString(path)
	builder.WriteString(" b/")
	builder.WriteString(path)
	builder.WriteString("\nnew file mode 100644\nindex 0000000..0000000\n--- /dev/null\n+++ b/")
	builder.WriteString(path)
	builder.WriteString("\n@@ -0,0 +1,")
	builder.WriteString(strconv.Itoa(len(lines)))
	builder.WriteString(" @@\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		builder.WriteString("+")
		builder.WriteString(strings.TrimSuffix(line, "\n"))
		builder.WriteString("\n")
	}

	return builder.String(), false
}

func countPatchLines(patch string) (int, int) {
	additions := 0
	deletions := 0

	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			deletions++
		}
	}

	return additions, deletions
}

func isBinaryPatch(patch string) bool {
	return strings.Contains(patch, "Binary files ") || strings.Contains(patch, "GIT binary patch")
}

func fingerprint(file ChangedFile) string {
	hash := sha1.New()
	hash.Write([]byte(file.Path))
	hash.Write([]byte(file.OldPath))
	hash.Write([]byte(file.Status))

	for _, section := range file.Sections {
		hash.Write([]byte(section.Kind))
		hash.Write([]byte(section.Patch))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func (file ChangedFile) pathSectionID(kind string) string {
	return fmt.Sprintf("%s:%s", kind, file.Path)
}

func listTrackedFiles(ctx context.Context, root string) ([]string, error) {
	raw, err := gitBytes(ctx, root, "ls-files", "-z")

	if err != nil {
		return nil, err
	}

	parts := bytes.Split(raw, []byte{0})
	files := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		files = append(files, string(part))
	}

	sort.Strings(files)

	return files, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := gitBytes(ctx, dir, args...)

	return string(output), err
}

func gitBytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		var exit *exec.ExitError

		if errors.As(err, &exit) {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exit.Stderr)))
		}

		return nil, err
	}

	return output, nil
}
