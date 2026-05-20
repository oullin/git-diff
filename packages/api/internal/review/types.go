package review

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
	// TrackedFiles contains every file under the repo root that is either tracked or
	// untracked-not-ignored. Ignored files are excluded.
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

// WorkingTreeDirtyError signals that a checkout was refused because the
// working tree has uncommitted changes that would be overwritten.
type WorkingTreeDirtyError struct {
	Files []string
}

type statusEntry struct {
	index  byte
	work   byte
	path   string
	old    string
	rename bool
}

type commitDiffEntry struct {
	status GitFileStatus
	path   string
	old    string
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

const maxFileReadBytes = 2 * 1024 * 1024

func (e *WorkingTreeDirtyError) Error() string {
	return "working tree has uncommitted changes"
}
