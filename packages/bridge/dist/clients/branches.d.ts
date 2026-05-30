import type { Branch, RepositoryState } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
export declare class BranchClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    list(request: { path?: string }): Promise<{
        branches: string[];
        records?: Branch[];
    }>;
    checkout(request: { path: string; branch: string }): Promise<RepositoryState>;
    create(request: { path: string; name: string }): Promise<RepositoryState>;
    delete(request: { path?: string; name: string }): Promise<void>;
    lock(request: { path?: string; name: string }): Promise<{
        branches: Branch[];
    }>;
    unlock(request: { path?: string; name: string }): Promise<{
        branches: Branch[];
    }>;
}
