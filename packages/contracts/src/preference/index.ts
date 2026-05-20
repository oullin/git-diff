export type DiffViewMode = "split" | "unified";
export type DiffHunkStyle = "soft" | "punchy" | "bar";
export type DiffDensity = "comfortable" | "compact";
export type UIAccent = "copper" | "indigo" | "emerald" | "amber" | "rose";

export interface UIPreferences {
  values: Record<string, string>;
  updatedAt?: string;
}

// Alias retained for back-compat with bridge transport naming.
export type UIPreferencesResponse = UIPreferences;

export const PREF_KEYS = {
  theme: "theme",
  diffViewMode: "diff.viewMode",
  diffHideWhitespace: "diff.hideWhitespace",
  diffStyle: "diff.style",
  diffDensity: "diff.density",
  diffWordHi: "diff.wordHi",
  lastRepoRoot: "repo.lastRoot",
  panelLeftWidth: "panel.left.width",
  panelFileTreeWidth: "panel.fileTree.width",
  panelRightWidth: "panel.right.width",
  uiAccent: "ui.accent",
  uiShowMinimap: "ui.minimap",
  uiShowStatusBar: "ui.statusBar",
} as const;

export const VIEWED_KEY_PREFIX = "file.viewed.";

export function viewedPrefKey(repoRoot: string, filePath: string): string {
  return `${VIEWED_KEY_PREFIX}${repoRoot}.${filePath}`;
}
