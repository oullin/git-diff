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
