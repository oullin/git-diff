export type LaunchIntentKind = "working" | "commit" | "pull-request" | "help";
export interface LaunchIntent {
    kind: LaunchIntentKind;
    repoPath?: string;
    /** Commit-ish to open. Accepts raw SHAs and HEAD/@ revision syntax. */
    commitRef?: string;
    /** @deprecated Use {@link commitRef}. Kept for one release. */
    sha?: string;
    pullRequestNumber?: number;
    /** @deprecated Use {@link pullRequestNumber}. Kept for one release. */
    prNumber?: number;
    /** Source GitHub URL when the user launched via `pr <url>`. */
    pullRequestUrl?: string;
    walkthrough: boolean;
    helpText?: string;
    raw?: string;
}
//# sourceMappingURL=index.d.ts.map
