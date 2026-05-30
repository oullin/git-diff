export const PREF_KEYS = {
    theme: "theme",
    diffViewMode: "diff.viewMode",
    diffHideWhitespace: "diff.hideWhitespace",
    diffStyle: "diff.style",
    diffDensity: "diff.density",
    diffWordHi: "diff.wordHi",
    diffHideResolved: "diff.hideResolved",
    lastRepoRoot: "repo.lastRoot",
    panelLeftWidth: "panel.left.width",
    panelFileTreeWidth: "panel.fileTree.width",
    panelRightWidth: "panel.right.width",
    uiAccent: "ui.accent",
    uiShowMinimap: "ui.minimap",
    uiShowStatusBar: "ui.statusBar",
};
export const VIEWED_KEY_PREFIX = "file.viewed.";
export function viewedPrefKey(repoRoot, filePath) {
    return `${VIEWED_KEY_PREFIX}${repoRoot}.${filePath}`;
}
