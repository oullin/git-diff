package setting

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateRuntimeSettings(home, fallbackRepo string, candidate RuntimeSettings) Validation {
	resolved := candidate.withDefaults(home, fallbackRepo)
	checks := []Check{}

	addCheck := func(key, label, path string, err error) {
		check := Check{Key: key, Label: label, Path: path, Status: CheckOK, Message: "ok"}

		if err != nil {
			check.Status = CheckError
			check.Message = err.Error()
		}

		checks = append(checks, check)
	}

	root, err := ValidateRepoRoot(resolved.RepoRoot)

	if err == nil {
		resolved.RepoRoot = root
		resolved = resolved.withDefaults(home, root)
	}

	addCheck("repo_root", "Repository root", resolved.RepoRoot, err)
	addCheck("workflow_db_path", "Review SQLite database", resolved.WorkflowDBPath, sqlitePathValid(resolved.WorkflowDBPath))

	valid := true

	for _, check := range checks {
		if check.Status != CheckOK {
			valid = false

			break
		}
	}

	return Validation{Settings: resolved, Checks: checks, Valid: valid}
}

func ValidateRepoRoot(root string) (string, error) {
	root = strings.TrimSpace(root)

	if root == "" {
		return "", errors.New("path is required")
	}

	abs, err := filepath.Abs(root)

	if err != nil {
		return root, err
	}

	if !HasRepoMarkers(abs) {
		return abs, fmt.Errorf("missing repository marker: expected %s or %s", filepath.Join(abs, ".git"), filepath.Join(abs, "go.mod"))
	}

	return abs, nil
}

func HasRepoMarkers(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))

	if err == nil && (info.IsDir() || !info.IsDir()) {
		return true
	}

	if info, err := os.Stat(filepath.Join(dir, "stow")); err == nil && info.IsDir() {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return true
		}
	}

	return false
}
