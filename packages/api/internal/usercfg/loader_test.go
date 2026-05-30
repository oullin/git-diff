package usercfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Loader{}.Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))

	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}

	want := Defaults()

	if cfg.Theme != want.Theme || cfg.Walkthrough.Provider != want.Walkthrough.Provider {
		t.Fatalf("missing-file load returned non-defaults: %+v", cfg)
	}
}

func TestLoaderPartialFileMergesDefaults(t *testing.T) {
	path := writeYAML(t, `
theme: dark
walkthrough:
  model: claude-sonnet-9
keymap:
  next_file: arrow-down
`)

	cfg, err := Loader{}.Load(path)

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Theme != "dark" {
		t.Fatalf("theme override missing: %q", cfg.Theme)
	}

	if cfg.Walkthrough.Model != "claude-sonnet-9" {
		t.Fatalf("walkthrough.model override missing: %q", cfg.Walkthrough.Model)
	}

	if cfg.Keymap.NextFile != "arrow-down" {
		t.Fatalf("keymap.next_file override missing: %q", cfg.Keymap.NextFile)
	}

	// Untouched keys keep defaults.
	if cfg.Walkthrough.Provider != Defaults().Walkthrough.Provider {
		t.Fatalf("provider should inherit default, got %q", cfg.Walkthrough.Provider)
	}

	if cfg.Keymap.PrevFile != Defaults().Keymap.PrevFile {
		t.Fatalf("prev_file should inherit default, got %q", cfg.Keymap.PrevFile)
	}
}

func TestLoaderRejectsMalformedYAML(t *testing.T) {
	path := writeYAML(t, "theme: dark\n  invalid: : :\n")

	_, err := Loader{}.Load(path)

	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoaderBudgetIntegerFields(t *testing.T) {
	path := writeYAML(t, `
walkthrough:
  patch_budget_bytes: 1024
  per_file_budget_bytes: 256
`)

	cfg, err := Loader{}.Load(path)

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Walkthrough.PatchBudgetBytes != 1024 || cfg.Walkthrough.PerFileBudgetBytes != 256 {
		t.Fatalf("budget overrides not applied: %+v", cfg.Walkthrough)
	}
}

func writeYAML(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	body = strings.TrimLeft(body, "\n")

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}
