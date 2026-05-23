package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/oullin/git-diff/internal/lines"
)

type PullRequestSummary struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	State   string `json:"state"`
	BaseRef string `json:"baseRef"`
	HeadRef string `json:"headRef"`
	URL     string `json:"url"`
}

// ErrGhUnavailable lets the HTTP layer surface a 412 when `gh` is missing.
var ErrGhUnavailable = errors.New("gh CLI is required for pull-request operations; install from https://cli.github.com")

// ListPullRequests returns open PRs (default 50, max 200).
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

// ReadPullRequestState anchors comments to head_sha so they stay attached
// when the PR branch updates upstream.
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

	// Fetch so the SHAs resolve locally if the user hasn't pulled this PR.
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

	return lines.SplitUnifiedPatch(raw), nil
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
