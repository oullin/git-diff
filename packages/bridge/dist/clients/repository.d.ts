import type {
    CommitSummary,
    RepositoryFile,
    RepositoryFileRange,
    RepositoryState,
} from "@git-diff/contracts";
import type { BytesResponse, HttpTransport } from "#bridge/http.js";
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
    /**
     * Fetch the raw bytes of a file at a given ref. Empty ref reads the
     * working tree, ":0" reads the index, otherwise treated as a git ref.
     * Returns the bytes plus the Content-Type so the renderer can build
     * an object URL for inline image display.
     */
    readFileBytes(request: { root?: string; path: string; ref?: string }): Promise<BytesResponse>;
    readFileRange(request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }): Promise<RepositoryFileRange>;
}
