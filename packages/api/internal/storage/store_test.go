package storage

import (
	"path/filepath"
	"testing"
)

func TestDefaultPathReturnsGitDiffNamespace(t *testing.T) {
	t.Setenv(envDBPath, "")

	home := t.TempDir()
	got := DefaultPath(home)
	want := filepath.Join(home, "Library", "Application Support", "git-diff", "reviews.sqlite3")

	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestDefaultPathHonorsEnvOverride(t *testing.T) {
	t.Setenv(envDBPath, "/tmp/custom.sqlite3")

	if got := DefaultPath(t.TempDir()); got != "/tmp/custom.sqlite3" {
		t.Fatalf("path = %q, want %q", got, "/tmp/custom.sqlite3")
	}
}
