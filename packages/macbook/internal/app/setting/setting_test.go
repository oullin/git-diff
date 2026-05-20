package setting

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRepoRootAcceptsExplicitValidRoot(t *testing.T) {
	repo := writeSettingsRepo(t)
	got, err := ValidateRepoRoot(repo)

	if err != nil {
		t.Fatal(err)
	}

	if got != repo {
		t.Fatalf("resolveRepoRoot = %q, want %q", got, repo)
	}
}

func TestResolveRepoRootRejectsExplicitInvalidRoot(t *testing.T) {
	dir := t.TempDir()

	if _, err := ValidateRepoRoot(dir); err == nil {
		t.Fatal("expected invalid repo root error")
	}
}

func TestValidateRuntimeSettingsResolvesPathsAndRejectsDirectoryDB(t *testing.T) {
	home := t.TempDir()
	repo := writeSettingsRepo(t)
	dbDir := filepath.Join(home, "dbdir")

	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatal(err)
	}

	validation := ValidateRuntimeSettings(home, repo, RuntimeSettings{
		RepoRoot:     repo,
		DatabasePath: dbDir,
	})

	if validation.Valid {
		t.Fatalf("expected invalid settings, got %#v", validation)
	}

	if !hasSettingsCheck(validation.Checks, "database_path", CheckError) {
		t.Fatalf("checks = %#v", validation.Checks)
	}
}

func TestValidateRuntimeSettingsAcceptsValidDefaults(t *testing.T) {
	home := t.TempDir()
	repo := writeSettingsRepo(t)
	validation := ValidateRuntimeSettings(home, repo, RuntimeSettings{RepoRoot: repo})

	if !validation.Valid {
		t.Fatalf("settings should be valid: %#v", validation.Checks)
	}

	if validation.Settings.DatabasePath == "" {
		t.Fatal("expected default database path")
	}
}

func writeSettingsRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()

	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}

	return repo
}

func hasSettingsCheck(checks []Check, key, status string) bool {
	for _, check := range checks {
		if check.Key == key && check.Status == status {
			return true
		}
	}

	return false
}
