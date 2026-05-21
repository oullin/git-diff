import type {
    CommitSummary,
    RepositoryFile,
    RepositoryFileRange,
    RepositoryState,
} from "@git-diff/contracts";
export declare function repositoryState(
    socketPath: string,
    request: {
        path?: string;
    },
): Promise<RepositoryState>;
export declare function openRepository(
    socketPath: string,
    request: {
        path: string;
    },
): Promise<RepositoryState>;
export declare function refreshRepository(
    socketPath: string,
    request: {
        path: string;
    },
): Promise<RepositoryState>;
export declare function readCommit(
    socketPath: string,
    request: {
        path?: string;
        sha: string;
    },
): Promise<RepositoryState>;
export declare function listCommits(
    socketPath: string,
    request: {
        path?: string;
        limit?: number;
    },
): Promise<{
    commits: CommitSummary[];
}>;
export declare function readRepositoryFile(
    socketPath: string,
    request: {
        root: string;
        path: string;
    },
): Promise<RepositoryFile>;
export declare function readRepositoryFileRange(
    socketPath: string,
    request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    },
): Promise<RepositoryFileRange>;
