package usercfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteDefaultsCreatesFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := WriteDefaults(path); err != nil {
		t.Fatalf("write: %v", err)
	}

	body, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	if !strings.Contains(string(body), "# git-diff user config") {
		t.Fatal("expected header comment in written file")
	}

	if !strings.Contains(string(body), "provider: anthropic") {
		t.Fatal("expected default walkthrough provider in written file")
	}
}

func TestWriteDefaultsIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := WriteDefaults(path); err != nil {
		t.Fatal(err)
	}

	// Modify the file; WriteDefaults must NOT overwrite.
	if err := os.WriteFile(path, []byte("theme: dark\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteDefaults(path); err != nil {
		t.Fatalf("second call: %v", err)
	}

	body, _ := os.ReadFile(path)

	if string(body) != "theme: dark\n" {
		t.Fatalf("WriteDefaults should not overwrite existing file; got %q", string(body))
	}
}

func TestWriteDefaultsCreatesMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "config.yaml")

	if err := WriteDefaults(path); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestWriteDefaultsRoundTripsThroughLoader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := WriteDefaults(path); err != nil {
		t.Fatal(err)
	}

	cfg, err := Loader{}.Load(path)

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := Defaults()

	if cfg.Theme != want.Theme {
		t.Fatalf("round-trip theme mismatch: %q vs %q", cfg.Theme, want.Theme)
	}

	if cfg.Walkthrough.PatchBudgetBytes != want.Walkthrough.PatchBudgetBytes {
		t.Fatalf("round-trip budget mismatch: %d vs %d",
			cfg.Walkthrough.PatchBudgetBytes, want.Walkthrough.PatchBudgetBytes)
	}

	if cfg.Keymap.CommandBar != want.Keymap.CommandBar {
		t.Fatalf("round-trip keymap mismatch: %q vs %q", cfg.Keymap.CommandBar, want.Keymap.CommandBar)
	}
}
