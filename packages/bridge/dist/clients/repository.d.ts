import type {
    CommitSummary,
    RepositoryFile,
    RepositoryFileRange,
    RepositoryState,
} from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class RepositoryClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    state(request: { path?: string }): Promise<RepositoryState>;
    open(request: { path: string }): Promise<RepositoryState>;
    refresh(request: { path: string }): Promise<RepositoryState>;
    readCommit(request: { path?: string; sha: string }): Promise<RepositoryState>;
    listCommits(request: { path?: string; limit?: number }): Promise<{
        commits: CommitSummary[];
    }>;
    readFile(request: { root: string; path: string }): Promise<RepositoryFile>;
    readFileRange(request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }): Promise<RepositoryFileRange>;
}
