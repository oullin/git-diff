import type { PullRequestSummary, RepositoryState } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class PullRequestClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    list(request: { path?: string; limit?: number }): Promise<{
        pullRequests: PullRequestSummary[];
    }>;
    read(request: { path?: string; number: number }): Promise<RepositoryState>;
}
