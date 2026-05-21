export type LaunchIntentKind = "working" | "commit" | "pull-request" | "help";

export interface LaunchIntent {
    kind: LaunchIntentKind;
    repoPath?: string;
    sha?: string;
    prNumber?: number;
    walkthrough: boolean;
    helpText?: string;
    raw?: string;
}
