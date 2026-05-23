package httpx

import (
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/usercfg"
)

func (s Server) userConfigGet(w http.ResponseWriter, _ *http.Request) {
	if s.UserConfig == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("user config not initialised"))

		return
	}

	writeJSON(w, http.StatusOK, marshalConfig(s.UserConfig.Get()))
}

// userConfigStream emits the current config immediately, then a "config"
// event on every YAML change — clients can subscribe without also calling
// userConfigGet.
func (s Server) userConfigStream(w http.ResponseWriter, r *http.Request) {
	if s.UserConfig == nil || s.UserConfigEvents == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("user config not initialised"))

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	if err := writeSSE(w, rc, "config", marshalConfig(s.UserConfig.Get())); err != nil {
		return
	}

	updates, unsubscribe := s.UserConfigEvents.Subscribe()

	defer unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case cfg, ok := <-updates:
			if !ok {
				return
			}

			if err := writeSSE(w, rc, "config", marshalConfig(cfg)); err != nil {
				return
			}
		}
	}
}

// marshalConfig emits the snake_case shape that mirrors the YAML schema —
// json marshalling the struct directly would leak the exported Go names.
func marshalConfig(cfg usercfg.Config) map[string]any {
	return map[string]any{
		"theme":                  cfg.Theme,
		"show_whitespace":        cfg.ShowWhitespace,
		"copy_comments_on_close": cfg.CopyCommentsOnClose,
		"last_repository_path":   cfg.LastRepositoryPath,
		"walkthrough": map[string]any{
			"provider":              cfg.Walkthrough.Provider,
			"model":                 cfg.Walkthrough.Model,
			"patch_budget_bytes":    cfg.Walkthrough.PatchBudgetBytes,
			"per_file_budget_bytes": cfg.Walkthrough.PerFileBudgetBytes,
		},
		"keymap": map[string]any{
			"command_bar":       cfg.Keymap.CommandBar,
			"file_filter":       cfg.Keymap.FileFilter,
			"diff_search":       cfg.Keymap.DiffSearch,
			"submit_comment":    cfg.Keymap.SubmitComment,
			"discard_comment":   cfg.Keymap.DiscardComment,
			"toggle_sidebar":    cfg.Keymap.ToggleSidebar,
			"next_file":         cfg.Keymap.NextFile,
			"prev_file":         cfg.Keymap.PrevFile,
			"next_hunk":         cfg.Keymap.NextHunk,
			"prev_hunk":         cfg.Keymap.PrevHunk,
			"toggle_viewed":     cfg.Keymap.ToggleViewed,
			"toggle_whitespace": cfg.Keymap.ToggleWhitespace,
		},
	}
}
