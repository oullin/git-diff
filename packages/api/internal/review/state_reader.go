package review

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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

	tracked, _ := ListRepositoryFiles(ctx, root)

	state := RepositoryState{
		Root:         root,
		LaunchPath:   launchPath,
		Mode:         RepositoryModeWorking,
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

func dirtyPaths(entries []statusEntry) []string {
	if len(entries) == 0 {
		return nil
	}

	paths := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.path == "" {
			continue
		}

		paths = append(paths, entry.path)
	}

	return paths
}
