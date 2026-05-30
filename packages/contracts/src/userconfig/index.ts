// Wire types for the user-editable YAML config served by the Go backend
// at GET /v1/userconfig. Keys are snake_case (per the YAML schema) so the
// renderer can swap between reading the file directly and the HTTP
// endpoint without translating.

export type ThemeChoice = "system" | "light" | "dark";

export type WalkthroughProvider = "anthropic" | "codex";

export interface WalkthroughConfig {
    provider: WalkthroughProvider;
    /** Model string the chosen provider supports. Providers validate. */
    model: string;
    patch_budget_bytes: number;
    per_file_budget_bytes: number;
}

/**
 * Action identifiers used as keys in the user-configurable keymap. The
 * Go backend mirrors this list — keep them in sync if either side adds
 * a new shortcut.
 */
export type KeymapAction =
    | "command_bar"
    | "file_filter"
    | "diff_search"
    | "submit_comment"
    | "discard_comment"
    | "toggle_sidebar"
    | "next_file"
    | "prev_file"
    | "next_hunk"
    | "prev_hunk"
    | "toggle_viewed"
    | "toggle_whitespace";

export type Keymap = Record<KeymapAction, string>;

export interface UserConfig {
    theme: ThemeChoice;
    show_whitespace: boolean;
    copy_comments_on_close: boolean;
    last_repository_path: string;
    walkthrough: WalkthroughConfig;
    keymap: Keymap;
    /** Absolute path to the YAML file (~/.git-diff/config.yaml). The
     *  renderer uses this to invoke shell.openPath without re-deriving it. */
    path: string;
}
