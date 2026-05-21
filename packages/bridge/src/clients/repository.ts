import type {
    CommitSummary,
    RepositoryFile,
    RepositoryFileRange,
    RepositoryState,
} from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

export class RepositoryClient {
    constructor(private readonly transport: HttpTransport) {}

    state(request: { path?: string }): Promise<RepositoryState> {
        const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";

        return this.transport.request<RepositoryState>("GET", `/v1/repository/state${query}`);
    }

    open(request: { path: string }): Promise<RepositoryState> {
        return this.transport.request<RepositoryState>("POST", "/v1/repository/open", {
            path: request.path,
        });
    }

    refresh(request: { path: string }): Promise<RepositoryState> {
        return this.transport.request<RepositoryState>("POST", "/v1/repository/refresh", {
            path: request.path,
        });
    }

    readCommit(request: { path?: string; sha: string }): Promise<RepositoryState> {
        const parts = [`sha=${encodeURIComponent(request.sha)}`];

        if (request.path) {
            parts.push(`path=${encodeURIComponent(request.path)}`);
        }

        return this.transport.request<RepositoryState>(
            "GET",
            `/v1/repository/commit?${parts.join("&")}`,
        );
    }

    listCommits(request: { path?: string; limit?: number }): Promise<{ commits: CommitSummary[] }> {
        const parts: string[] = [];

        if (request.path) {
            parts.push(`path=${encodeURIComponent(request.path)}`);
        }

        if (request.limit) {
            parts.push(`limit=${request.limit}`);
        }

        const query = parts.length === 0 ? "" : `?${parts.join("&")}`;

        return this.transport.request<{ commits: CommitSummary[] }>(
            "GET",
            `/v1/repository/log${query}`,
        );
    }

    readFile(request: { root: string; path: string }): Promise<RepositoryFile> {
        const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;

        return this.transport.request<RepositoryFile>("GET", `/v1/repository/file${query}`);
    }

    readFileRange(request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }): Promise<RepositoryFileRange> {
        const parts = [
            `root=${encodeURIComponent(request.root)}`,
            `path=${encodeURIComponent(request.path)}`,
            `startLine=${request.startLine}`,
            `endLine=${request.endLine}`,
        ];

        if (request.ref) {
            parts.push(`ref=${encodeURIComponent(request.ref)}`);
        }

        return this.transport.request<RepositoryFileRange>(
            "GET",
            `/v1/repository/file-range?${parts.join("&")}`,
        );
    }
}
