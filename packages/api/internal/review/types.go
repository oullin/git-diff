package review

import "github.com/oullin/git-diff/internal/domain/repostate"

// Domain model lives in internal/domain/repostate. These aliases keep existing
// review.* references compiling while consumers migrate to the domain package.
type (
	GitFileStatus       = repostate.GitFileStatus
	DiffSection         = repostate.DiffSection
	ChangedFile         = repostate.ChangedFile
	RepositoryState     = repostate.RepositoryState
	CommitSummary       = repostate.CommitSummary
	RepositoryFile      = repostate.RepositoryFile
	RepositoryFileRange = repostate.RepositoryFileRange
)

// Git-adapter internals stay in the review package.

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
	StatusAdded     = repostate.StatusAdded
	StatusDeleted   = repostate.StatusDeleted
	StatusModified  = repostate.StatusModified
	StatusRenamed   = repostate.StatusRenamed
	StatusUntracked = repostate.StatusUntracked
)

const (
	RepositoryModeWorking = repostate.RepositoryModeWorking
	RepositoryModeCommit  = repostate.RepositoryModeCommit
)

const maxFileReadBytes = 2 * 1024 * 1024

// maxFileRangeLines: clients asking for more must paginate.
const maxFileRangeLines = 500

func (e *WorkingTreeDirtyError) Error() string {
	return "working tree has uncommitted changes"
}
