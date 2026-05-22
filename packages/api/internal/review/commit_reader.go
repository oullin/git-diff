package review

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/lineparse"
)

func ReadCommitState(ctx context.Context, launchPath, sha string) (RepositoryState, error) {
	if strings.TrimSpace(sha) == "" {
		return RepositoryState{}, errors.New("commit sha is required")
	}

	ctx = WithRootCache(ctx)

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return RepositoryState{}, err
	}

	resolved, err := ResolveCommitRef(ctx, root, sha)

	if err != nil {
		return RepositoryState{}, fmt.Errorf("verify commit: %w", err)
	}

	short, _ := gitOutput(ctx, root, "rev-parse", "--short", resolved)
	short = strings.TrimSpace(short)

	nameStatusRaw, err := gitBytes(ctx, root, "diff-tree", "--no-commit-id", "--name-status", "-z", "-r", "--find-renames", "--root", resolved)

	if err != nil {
		return RepositoryState{}, fmt.Errorf("list commit files: %w", err)
	}

	entries := parseDiffTreeNameStatus(nameStatusRaw)

	// Batch: one `git show` covers every file in the commit. Splitting it
	// gives us per-file bodies without the N+1 invocation cost the old code
	// paid when commits touched many files.
	patches, err := readCommitPatches(ctx, root, resolved)

	if err != nil {
		return RepositoryState{}, err
	}

	files := buildCommitChangedFiles(entries, patches, resolved)

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

func readCommitPatches(ctx context.Context, root, sha string) (map[string][]byte, error) {
	raw, err := gitBytes(ctx, root,
		"show", "--binary", "--find-renames", "--format=", "-m", "--first-parent", sha,
	)

	if err != nil {
		return nil, fmt.Errorf("read commit patches: %w", err)
	}

	return lineparse.SplitUnifiedPatch(raw), nil
}

func buildCommitChangedFiles(entries []commitDiffEntry, patches map[string][]byte, sha string) []ChangedFile {
	files := make([]ChangedFile, 0, len(entries))

	for _, entry := range entries {
		patch := string(patches[entry.path])
		binary := isBinaryPatch(patch)

		section := DiffSection{
			ID:     fmt.Sprintf("commit:%s:%s", sha, entry.path),
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

	return files
}
