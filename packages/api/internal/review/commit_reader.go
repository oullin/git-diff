package review

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

func ReadCommitState(ctx context.Context, launchPath, sha string) (RepositoryState, error) {
	if strings.TrimSpace(sha) == "" {
		return RepositoryState{}, errors.New("commit sha is required")
	}

	root, err := gitOutput(ctx, launchPath, "rev-parse", "--show-toplevel")

	if err != nil {
		return RepositoryState{}, fmt.Errorf("resolve git root: %w", err)
	}

	root = strings.TrimSpace(root)

	resolved, err := gitOutput(ctx, root, "rev-parse", "--verify", sha+"^{commit}")

	if err != nil {
		return RepositoryState{}, fmt.Errorf("verify commit: %w", err)
	}

	resolved = strings.TrimSpace(resolved)
	short, _ := gitOutput(ctx, root, "rev-parse", "--short", resolved)
	short = strings.TrimSpace(short)

	nameStatusRaw, err := gitBytes(ctx, root, "diff-tree", "--no-commit-id", "--name-status", "-z", "-r", "--find-renames", "--root", resolved)

	if err != nil {
		return RepositoryState{}, fmt.Errorf("list commit files: %w", err)
	}

	entries := parseDiffTreeNameStatus(nameStatusRaw)
	files := make([]ChangedFile, 0, len(entries))

	for _, entry := range entries {
		patch, binary := commitFilePatch(ctx, root, resolved, entry.path, entry.old)

		section := DiffSection{
			ID:     fmt.Sprintf("commit:%s:%s", resolved, entry.path),
			Kind:   "commit",
			Patch:  patch,
			Binary: binary,
		}

		file := ChangedFile{
			Path:     entry.path,
			OldPath:  entry.old,
			Status:   entry.status,
			Binary:   binary,
			Sections: []DiffSection{section},
		}

		file.Additions, file.Deletions = countPatchLines(patch)
		file.Fingerprint = fingerprint(file)
		files = append(files, file)
	}

	sort.SliceStable(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	tracked, _ := ListRepositoryFiles(ctx, root)

	state := RepositoryState{
		Root:         root,
		LaunchPath:   launchPath,
		Mode:         RepositoryModeCommit,
		Branch:       short,
		HeadSHA:      resolved,
		CommitSHA:    resolved,
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

func parseDiffTreeNameStatus(raw []byte) []commitDiffEntry {
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

func commitFilePatch(ctx context.Context, root, sha, path, oldPath string) (string, bool) {
	args := []string{"show", "--binary", "--find-renames", "--format=", "-m", "--first-parent", sha, "--"}

	if oldPath != "" {
		args = append(args, oldPath)
	}

	args = append(args, path)
	patch, err := gitOutput(ctx, root, args...)

	if err != nil {
		return "", false
	}

	return patch, isBinaryPatch(patch)
}
