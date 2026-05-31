// Package repostate holds the repository-state domain model: the pure data
// types describing a repository's diff/commit state. It has no I/O and no
// dependency on git execution — the review package (git adapter) produces these
// types, and the httpx/service/walks layers consume them.
package repostate

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
	Root        string        `json:"root"`
	LaunchPath  string        `json:"launchPath"`
	Mode        string        `json:"mode"`
	Branch      string        `json:"branch"`
	HeadSHA     string        `json:"headSha"`
	CommitSHA   string        `json:"commitSha,omitempty"`
	GeneratedAt string        `json:"generatedAt"`
	Files       []ChangedFile `json:"files"`
	// TrackedFiles excludes ignored files (gitignore'd paths don't appear).
	TrackedFiles []string `json:"trackedFiles"`
	Additions    int      `json:"additions"`
	Deletions    int      `json:"deletions"`
}

type CommitSummary struct {
	SHA      string `json:"sha"`
	ShortSHA string `json:"shortSha"`
	Author   string `json:"author"`
	Email    string `json:"email"`
	Date     string `json:"date"`
	Subject  string `json:"subject"`
}

type RepositoryFile struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Binary    bool   `json:"binary"`
	Truncated bool   `json:"truncated"`
	Size      int64  `json:"size"`
}

type RepositoryFileRange struct {
	Path      string   `json:"path"`
	StartLine int      `json:"startLine"`
	EndLine   int      `json:"endLine"`
	Lines     []string `json:"lines"`
	EOF       bool     `json:"eof"`
}

const (
	StatusAdded     GitFileStatus = "added"
	StatusDeleted   GitFileStatus = "deleted"
	StatusModified  GitFileStatus = "modified"
	StatusRenamed   GitFileStatus = "renamed"
	StatusUntracked GitFileStatus = "untracked"
)

const (
	RepositoryModeWorking = "working"
	RepositoryModeCommit  = "commit"
)
