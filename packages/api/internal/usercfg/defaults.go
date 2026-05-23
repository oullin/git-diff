package usercfg

// Defaults must always be a complete, valid Config so the app boots even
// when no YAML file exists.
func Defaults() Config {
	return Config{
		Theme:               "system",
		ShowWhitespace:      false,
		CopyCommentsOnClose: false,
		LastRepositoryPath:  "",
		Walkthrough: Walkthrough{
			Provider:           "anthropic",
			Model:              "claude-sonnet-4-5",
			PatchBudgetBytes:   160 * 1024,
			PerFileBudgetBytes: 4 * 1024,
		},
		Anthropic: Anthropic{
			Endpoint:         "https://api.anthropic.com/v1/messages",
			APIVersion:       "2023-06-01",
			DefaultModel:     "claude-sonnet-4-5",
			ErrorBodyLimit:   64 * 1024,
			SuccessBodyLimit: 8 * 1024 * 1024,
		},
		Codex: Codex{
			DefaultModel: "gpt-5.3-codex-spark",
		},
		Keymap: Keymap{
			CommandBar:       "cmd+shift+p",
			FileFilter:       "cmd+f",
			DiffSearch:       "/",
			SubmitComment:    "cmd+enter",
			DiscardComment:   "escape",
			ToggleSidebar:    "cmd+\\",
			NextFile:         "j",
			PrevFile:         "k",
			NextHunk:         "n",
			PrevHunk:         "p",
			ToggleViewed:     "v",
			ToggleWhitespace: "w",
		},
	}
}
