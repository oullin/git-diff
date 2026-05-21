import type { PullRequestSummary, RepositoryState } from "@git-diff/contracts";
export declare function listPullRequests(
    socketPath: string,
    request: {
        path?: string;
        limit?: number;
    },
): Promise<{
    pullRequests: PullRequestSummary[];
}>;
export declare function readPullRequest(
    socketPath: string,
    request: {
        path?: string;
        number: number;
    },
): Promise<RepositoryState>;
