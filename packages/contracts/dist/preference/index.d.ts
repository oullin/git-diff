export type DiffViewMode = "split" | "unified";
export type DiffHunkStyle = "soft" | "punchy" | "bar";
export type DiffDensity = "comfortable" | "compact";
export type UIAccent = "copper" | "indigo" | "emerald" | "amber" | "rose";
export interface UIPreferences {
    values: Record<string, string>;
    updatedAt?: string;
}
export type UIPreferencesResponse = UIPreferences;
export declare const PREF_KEYS: {
    readonly theme: "theme";
    readonly diffViewMode: "diff.viewMode";
    readonly diffHideWhitespace: "diff.hideWhitespace";
    readonly diffStyle: "diff.style";
    readonly diffDensity: "diff.density";
    readonly diffWordHi: "diff.wordHi";
    readonly lastRepoRoot: "repo.lastRoot";
    readonly panelLeftWidth: "panel.left.width";
    readonly panelFileTreeWidth: "panel.fileTree.width";
    readonly panelRightWidth: "panel.right.width";
    readonly uiAccent: "ui.accent";
    readonly uiShowMinimap: "ui.minimap";
    readonly uiShowStatusBar: "ui.statusBar";
};
export declare const VIEWED_KEY_PREFIX = "file.viewed.";
export declare function viewedPrefKey(repoRoot: string, filePath: string): string;
//# sourceMappingURL=index.d.ts.map