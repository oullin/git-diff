package command

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureSystemPath prepends Homebrew and user-local bin directories so
// launchd-spawned processes can find them. No-op outside darwin.
func EnsureSystemPath() {
	if runtime.GOOS != "darwin" {
		return
	}

	candidates := []string{
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, "bin"),
		)
	}

	updated := augmentPath(os.Getenv("PATH"), candidates, isDir)

	if updated == "" {
		return
	}

	os.Setenv("PATH", updated)
}

// augmentPath returns "" when no changes are needed.
func augmentPath(current string, candidates []string, exists func(string) bool) string {
	existing := make(map[string]struct{})

	for _, entry := range strings.Split(current, string(os.PathListSeparator)) {
		if entry != "" {
			existing[entry] = struct{}{}
		}
	}

	var prepend []string

	for _, dir := range candidates {
		if _, seen := existing[dir]; seen {
			continue
		}

		if !exists(dir) {
			continue
		}

		prepend = append(prepend, dir)
		existing[dir] = struct{}{}
	}

	if len(prepend) == 0 {
		return ""
	}

	updated := strings.Join(prepend, string(os.PathListSeparator))

	if current != "" {
		updated = updated + string(os.PathListSeparator) + current
	}

	return updated
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}
