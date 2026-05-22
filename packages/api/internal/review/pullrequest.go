package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gocanto/git-diff/internal/lineparse"
)

// PullRequestSummary describes one entry returned by `gh pr list`. The UI
// renders these in a sidebar tab so the user can pick a PR to review.
type PullRequestSummary struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	State   string `json:"state"`
	BaseRef string `json:"baseRef"`
	HeadRef string `json:"headRef"`
	URL     string `json:"url"`
}

// ErrGhUnavailable is returned when the `gh` CLI isn't on PATH. The HTTP
// layer maps this to a 412 so the UI can prompt the user to install it.
var ErrGhUnavailable = errors.New("gh CLI is required for pull-request operations; install from https://cli.github.com")

// ListPullRequests returns the open pull requests for the repository at
// launchPath. Backed by `gh pr list --json`. Limited to 50 entries — enough
// to scroll through but cheap to fetch.
func ListPullRequests(ctx context.Context, launchPath string, limit int) ([]PullRequestSummary, error) {
	if !hasGh(ctx) {
		return nil, ErrGhUnavailable
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}

	raw, err := ghOutput(ctx, root,
		"pr", "list",
		"--state", "open",
		"--limit", strconv.Itoa(limit),
		"--json", "number,title,author,state,baseRefName,headRefName,url",
	)

	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}

	type ghAuthor struct {
		Login string `json:"login"`
	}

	type ghPR struct {
		Number      int      `json:"number"`
		Title       string   `json:"title"`
		Author      ghAuthor `json:"author"`
		State       string   `json:"state"`
		BaseRefName string   `json:"baseRefName"`
		HeadRefName string   `json:"headRefName"`
		URL         string   `json:"url"`
	}

	var rows []ghPR

	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("decode gh pr list output: %w", err)
	}

	summaries := make([]PullRequestSummary, 0, len(rows))

	for _, row := range rows {
		summaries = append(summaries, PullRequestSummary{
			Number:  row.Number,
			Title:   row.Title,
			Author:  row.Author.Login,
			State:   row.State,
			BaseRef: row.BaseRefName,
			HeadRef: row.HeadRefName,
			URL:     row.URL,
		})
	}

	return summaries, nil
}

// ReadPullRequestState fetches the PR head into a local ref and renders the
// base..head diff through the same shape as ReadCommitState. Comments
// attach to commit SHAs (head_sha) so they stay anchored even if the PR
// branch updates upstream.
func ReadPullRequestState(ctx context.Context, launchPath string, number int) (RepositoryState, error) {
	if number <= 0 {
		return RepositoryState{}, errors.New("pull request number is required")
	}

	if !hasGh(ctx) {
		return RepositoryState{}, ErrGhUnavailable
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return RepositoryState{}, err
	}

	detail, err := ghOutput(ctx, root,
		"pr", "view", strconv.Itoa(number),
		"--json", "title,author,baseRefName,headRefName,baseRefOid,headRefOid,url",
	)

	if err != nil {
		return RepositoryState{}, fmt.Errorf("gh pr view: %w", err)
	}

	type ghAuthor struct {
		Login string `json:"login"`
	}

	type ghView struct {
		Title       string   `json:"title"`
		Author      ghAuthor `json:"author"`
		BaseRefName string   `json:"baseRefName"`
		HeadRefName string   `json:"headRefName"`
		BaseRefOid  string   `json:"baseRefOid"`
		HeadRefOid  string   `json:"headRefOid"`
		URL         string   `json:"url"`
	}

	var view ghView

	if err := json.Unmarshal(detail, &view); err != nil {
		return RepositoryState{}, fmt.Errorf("decode gh pr view: %w", err)
	}

	// Fetch the head ref so the SHAs resolve locally even if the user hasn't
	// pulled this PR before.
	if _, err := gitOutput(ctx, root, "fetch", "origin", fmt.Sprintf("pull/%d/head", number)); err != nil {
		return RepositoryState{}, fmt.Errorf("fetch PR head: %w", err)
	}

	nameStatusRaw, err := gitBytes(ctx, root,
		"diff", "--name-status", "-z", "--find-renames",
		view.BaseRefOid+".."+view.HeadRefOid,
	)

	if err != nil {
		return RepositoryState{}, fmt.Errorf("list PR files: %w", err)
	}

	entries := parseDiffTreeNameStatus(nameStatusRaw)

	// One batched `git diff` covers every file in the PR; we split by
	// `diff --git` headers so each file's patch is recovered without the
	// N+1 invocation cost the per-file loop used to pay.
	patches, err := readPullRequestPatches(ctx, root, view.BaseRefOid, view.HeadRefOid)

	if err != nil {
		return RepositoryState{}, err
	}

	files := buildPullRequestChangedFiles(entries, patches, number)

	sort.SliceStable(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	tracked, _ := ListRepositoryFiles(ctx, root)

	state := RepositoryState{
		Root:         root,
		LaunchPath:   launchPath,
		Mode:         RepositoryModeCommit,
		Branch:       fmt.Sprintf("PR #%d (%s)", number, view.HeadRefName),
		HeadSHA:      view.HeadRefOid,
		CommitSHA:    view.HeadRefOid,
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

func readPullRequestPatches(ctx context.Context, root, baseSHA, headSHA string) (map[string][]byte, error) {
	raw, err := gitBytes(ctx, root,
		"diff", "--binary", "--find-renames",
		baseSHA+".."+headSHA,
	)

	if err != nil {
		return nil, fmt.Errorf("read PR patches: %w", err)
	}

	return lineparse.SplitUnifiedPatch(raw), nil
}

func buildPullRequestChangedFiles(entries []commitDiffEntry, patches map[string][]byte, number int) []ChangedFile {
	files := make([]ChangedFile, 0, len(entries))

	for _, entry := range entries {
		patch := string(patches[entry.path])
		binary := isBinaryPatch(patch)

		section := DiffSection{
			ID:     fmt.Sprintf("pr:%d:%s", number, entry.path),
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
