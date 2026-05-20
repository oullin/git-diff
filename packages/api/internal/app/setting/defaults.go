package setting

import (
	"strings"

	"github.com/gocanto/git-diff/internal/storage"
)

func DefaultRuntimeSettings(home, repo string) RuntimeSettings {
	return RuntimeSettings{
		RepoRoot:     repo,
		DatabasePath: storage.DefaultPath(home),
	}
}

func (s RuntimeSettings) withDefaults(home, fallbackRepo string) RuntimeSettings {
	if strings.TrimSpace(s.RepoRoot) == "" {
		s.RepoRoot = fallbackRepo
	}

	repo := resolvePath(home, fallbackRepo, s.RepoRoot)
	s.RepoRoot = repo

	defaults := DefaultRuntimeSettings(home, repo)

	if strings.TrimSpace(s.DatabasePath) == "" {
		s.DatabasePath = defaults.DatabasePath
	}

	s.DatabasePath = resolvePath(home, repo, s.DatabasePath)

	return s
}
