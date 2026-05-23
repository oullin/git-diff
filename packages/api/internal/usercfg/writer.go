package usercfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WriteDefaults is idempotent: existing files are left alone, so
// callers can invoke this unconditionally at startup.
func WriteDefaults(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	body, err := yaml.Marshal(Defaults())

	if err != nil {
		return fmt.Errorf("marshal defaults: %w", err)
	}

	contents := append([]byte(headerComment), body...)

	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

const headerComment = `# git-diff user config — edit and save; changes hot-reload.
# Schema: https://github.com/oullin/git-diff#config
#
# walkthrough.provider:        anthropic | codex
# walkthrough.model:           model string the chosen provider supports
# theme:                       system | light | dark
# anthropic.endpoint:          messages API URL (override for proxies)
# anthropic.api_version:       value sent in the anthropic-version header
# anthropic.default_model:     fallback model when walkthrough.model is empty
# anthropic.error_body_limit:  max bytes read from non-2xx responses
# anthropic.success_body_limit: max bytes read from successful responses
# codex.default_model:         fallback model when walkthrough.model is empty
# keymap.*:                    "modifier+...+key" strings (cmd, ctrl, shift, alt + a single key)
#
`
