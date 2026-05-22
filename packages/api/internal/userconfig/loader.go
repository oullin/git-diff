package userconfig

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Loader reads a YAML config from disk and decodes into Config. Stateless
// — Load is safe to call concurrently with distinct paths.
//
// Defaults are merged per-key via viper.SetDefault so a partial YAML file
// inherits the missing keys from Defaults() automatically.
type Loader struct{}

// Load reads and decodes the config at path. Missing files are treated as
// "use defaults" — the caller decides whether to call Writer.WriteDefaults
// to scaffold one.
func (Loader) Load(path string) (Config, error) {
	v := viper.New()

	applyDefaults(v, Defaults())

	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		if isMissing(err) {
			cfg := Defaults()

			return cfg, nil
		}

		return Config{}, fmt.Errorf("read user config %s: %w", path, err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode user config: %w", err)
	}

	return cfg, nil
}

func isMissing(err error) bool {
	var notFound viper.ConfigFileNotFoundError

	if errors.As(err, &notFound) {
		return true
	}

	return errors.Is(err, os.ErrNotExist)
}

// applyDefaults populates viper's default layer from a Defaults() snapshot
// so partial files inherit missing keys per-leaf, not all-or-nothing.
func applyDefaults(v *viper.Viper, d Config) {
	v.SetDefault("theme", d.Theme)
	v.SetDefault("show_whitespace", d.ShowWhitespace)
	v.SetDefault("copy_comments_on_close", d.CopyCommentsOnClose)
	v.SetDefault("last_repository_path", d.LastRepositoryPath)

	v.SetDefault("walkthrough.provider", d.Walkthrough.Provider)
	v.SetDefault("walkthrough.model", d.Walkthrough.Model)
	v.SetDefault("walkthrough.patch_budget_bytes", d.Walkthrough.PatchBudgetBytes)
	v.SetDefault("walkthrough.per_file_budget_bytes", d.Walkthrough.PerFileBudgetBytes)

	v.SetDefault("keymap.command_bar", d.Keymap.CommandBar)
	v.SetDefault("keymap.file_filter", d.Keymap.FileFilter)
	v.SetDefault("keymap.diff_search", d.Keymap.DiffSearch)
	v.SetDefault("keymap.submit_comment", d.Keymap.SubmitComment)
	v.SetDefault("keymap.discard_comment", d.Keymap.DiscardComment)
	v.SetDefault("keymap.toggle_sidebar", d.Keymap.ToggleSidebar)
	v.SetDefault("keymap.next_file", d.Keymap.NextFile)
	v.SetDefault("keymap.prev_file", d.Keymap.PrevFile)
	v.SetDefault("keymap.next_hunk", d.Keymap.NextHunk)
	v.SetDefault("keymap.prev_hunk", d.Keymap.PrevHunk)
	v.SetDefault("keymap.toggle_viewed", d.Keymap.ToggleViewed)
	v.SetDefault("keymap.toggle_whitespace", d.Keymap.ToggleWhitespace)
}
