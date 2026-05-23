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
export type KeymapAction = "command_bar" | "file_filter" | "diff_search" | "submit_comment" | "discard_comment" | "toggle_sidebar" | "next_file" | "prev_file" | "next_hunk" | "prev_hunk" | "toggle_viewed" | "toggle_whitespace";
export type Keymap = Record<KeymapAction, string>;
export interface UserConfig {
    theme: ThemeChoice;
    show_whitespace: boolean;
    copy_comments_on_close: boolean;
    last_repository_path: string;
    walkthrough: WalkthroughConfig;
    keymap: Keymap;
}
//# sourceMappingURL=index.d.ts.map