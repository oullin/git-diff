package httpx

import (
	"errors"
	"net/http"

	"github.com/gocanto/git-diff/internal/userconfig"
)

// userConfigGet returns the current resolved config as JSON. Snake_case
// keys are preserved via the struct's yaml tags + a hand-built map below
// (json marshalling would otherwise default to the exported Go names).
func (s Server) userConfigGet(w http.ResponseWriter, _ *http.Request) {
	if s.UserConfig == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("user config not initialised"))

		return
	}

	writeJSON(w, http.StatusOK, marshalConfig(s.UserConfig.Get()))
}

// userConfigStream is an SSE endpoint that emits a "config" event whenever
// the YAML file changes. The initial event fires immediately so clients
// can subscribe and forget about the GET endpoint.
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

// marshalConfig converts a Config into a snake_case map matching the YAML
// schema 1:1, so the renderer can consume the same key names whether it
// reads the YAML directly or our HTTP endpoint. Centralised here so the
// JSON shape lives next to the handler that emits it.
func marshalConfig(cfg userconfig.Config) map[string]any {
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
