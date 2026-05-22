// Package userconfig owns the user-editable YAML config at
// ~/.git-diff/config.yaml. Decomposed so each concern lives in its own
// file: schema (data), defaults (pure function), loader (viper wiring),
// watcher (notify on edits), writer (first-run scaffold), reader (the
// interface handlers depend on).
package userconfig

// Config is the in-memory shape of the user config. Snake_case wire format
// per the user's decision; `mapstructure` tags are how viper drives the
// decode from YAML → Config.
//
// All fields are value types so passing a Config across goroutines is safe
// without locking — callers either hold their own copy or read through the
// Reader interface.
type Config struct {
	Theme               string      `mapstructure:"theme"               yaml:"theme"`
	ShowWhitespace      bool        `mapstructure:"show_whitespace"     yaml:"show_whitespace"`
	CopyCommentsOnClose bool        `mapstructure:"copy_comments_on_close" yaml:"copy_comments_on_close"`
	LastRepositoryPath  string      `mapstructure:"last_repository_path" yaml:"last_repository_path"`
	Walkthrough         Walkthrough `mapstructure:"walkthrough"         yaml:"walkthrough"`
	Keymap              Keymap      `mapstructure:"keymap"              yaml:"keymap"`
}

// Walkthrough groups the LLM-walkthrough settings. The model field is a
// single string interpreted by whichever provider is selected — providers
// validate and reject models they don't support.
type Walkthrough struct {
	Provider           string `mapstructure:"provider"               yaml:"provider"`
	Model              string `mapstructure:"model"                  yaml:"model"`
	PatchBudgetBytes   int    `mapstructure:"patch_budget_bytes"     yaml:"patch_budget_bytes"`
	PerFileBudgetBytes int    `mapstructure:"per_file_budget_bytes"  yaml:"per_file_budget_bytes"`
}

// Keymap is the user-configurable shortcut table. Every action is a string
// in the "modifier+...+key" format consumed by the renderer's key matcher
// (e.g. "cmd+shift+p", "ctrl+enter", "/"). Empty strings disable the
// action.
type Keymap struct {
	CommandBar       string `mapstructure:"command_bar"        yaml:"command_bar"`
	FileFilter       string `mapstructure:"file_filter"        yaml:"file_filter"`
	DiffSearch       string `mapstructure:"diff_search"        yaml:"diff_search"`
	SubmitComment    string `mapstructure:"submit_comment"     yaml:"submit_comment"`
	DiscardComment   string `mapstructure:"discard_comment"    yaml:"discard_comment"`
	ToggleSidebar    string `mapstructure:"toggle_sidebar"     yaml:"toggle_sidebar"`
	NextFile         string `mapstructure:"next_file"          yaml:"next_file"`
	PrevFile         string `mapstructure:"prev_file"          yaml:"prev_file"`
	NextHunk         string `mapstructure:"next_hunk"          yaml:"next_hunk"`
	PrevHunk         string `mapstructure:"prev_hunk"          yaml:"prev_hunk"`
	ToggleViewed     string `mapstructure:"toggle_viewed"      yaml:"toggle_viewed"`
	ToggleWhitespace string `mapstructure:"toggle_whitespace"  yaml:"toggle_whitespace"`
}
