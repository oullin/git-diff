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

	"github.com/oullin/git-diff/internal/lines"
	"golang.org/x/sync/errgroup"
)

func ReadRepositoryState(ctx context.Context, launchPath string) (RepositoryState, error) {
	ctx = WithRootCache(ctx)

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return RepositoryState{}, err
	}

	// Four independent git queries fan out so wall time collapses to the
	// slowest one rather than summing across all four.
	var (
		branch    string
		head      string
		statusRaw []byte
		staged    map[string][]byte
		unstaged  map[string][]byte
	)

	group, gctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		out, _ := gitOutput(gctx, root, "branch", "--show-current")
		branch = strings.TrimSpace(out)

		return nil
	})

	group.Go(func() error {
		out, _ := gitOutput(gctx, root, "rev-parse", "--short", "HEAD")
		head = strings.TrimSpace(out)

		return nil
	})

	group.Go(func() error {
		out, err := gitBytes(gctx, root, "status", "--porcelain=v1", "-z")

		if err != nil {
			return fmt.Errorf("read git status: %w", err)
		}

		statusRaw = out

		return nil
	})

	group.Go(func() error {
		raw, err := gitBytes(gctx, root, "diff", "--binary", "--find-renames", "--cached")

		if err != nil {
			return fmt.Errorf("read staged diff: %w", err)
		}

		staged = lines.SplitUnifiedPatch(raw)

		return nil
	})

	group.Go(func() error {
		raw, err := gitBytes(gctx, root, "diff", "--binary", "--find-renames")

		if err != nil {
			return fmt.Errorf("read unstaged diff: %w", err)
		}

		unstaged = lines.SplitUnifiedPatch(raw)

		return nil
	})

	if err := group.Wait(); err != nil {
		return RepositoryState{}, err
	}

	entries := parseStatus(statusRaw)
	files := assembleChangedFiles(entries, root, staged, unstaged)

	sort.SliceStable(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	tracked, _ := ListRepositoryFiles(ctx, root)

	state := RepositoryState{
		Root:         root,
		LaunchPath:   launchPath,
		Mode:         RepositoryModeWorking,
		Branch:       branch,
		HeadSHA:      head,
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

// assembleChangedFiles synthesises a green-line patch for untracked files
// since they're absent from both staged and unstaged diff outputs.
func assembleChangedFiles(
	entries []statusEntry,
	root string,
	staged, unstaged map[string][]byte,
) []ChangedFile {
	files := make([]ChangedFile, 0, len(entries))

	for _, entry := range entries {
		file := ChangedFile{
			Path:    entry.path,
			OldPath: entry.old,
			Status:  statusFromEntry(entry),
		}

		if entry.index != ' ' && entry.index != '?' {
			patch := string(staged[entry.path])
			binary := isBinaryPatch(patch)
			file.Sections = append(file.Sections, DiffSection{ID: pathSectionID(file, "staged"), Kind: "staged", Patch: patch, Binary: binary})
			file.Binary = file.Binary || binary
			add, del := countPatchLines(patch)
			file.Additions += add
			file.Deletions += del
		}

		if entry.work != ' ' && entry.work != '?' {
			patch := string(unstaged[entry.path])
			binary := isBinaryPatch(patch)
			file.Sections = append(file.Sections, DiffSection{ID: pathSectionID(file, "unstaged"), Kind: "unstaged", Patch: patch, Binary: binary})
			file.Binary = file.Binary || binary
			add, del := countPatchLines(patch)
			file.Additions += add
			file.Deletions += del
		}

		if entry.index == '?' && entry.work == '?' {
			patch, binary := untrackedPatch(root, entry.path)
			file.Sections = append(file.Sections, DiffSection{ID: pathSectionID(file, "untracked"), Kind: "untracked", Patch: patch, Binary: binary})
			file.Binary = binary
			file.Additions, file.Deletions = countPatchLines(patch)
		}

		if len(file.Sections) == 0 {
			continue
		}

		file.Fingerprint = fingerprint(file)
		files = append(files, file)
	}

	return files
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
