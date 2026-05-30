import type { FileSearchResult, Repository, RepositoryCollaborator } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
export declare class RepositoriesClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    list(): Promise<{
        repositories: Repository[];
    }>;
    searchFiles(request: { query: string; limit?: number }): Promise<{
        results: FileSearchResult[];
    }>;
    upsert(request: { path: string; name?: string }): Promise<Repository>;
    remove(request: { path: string }): Promise<void>;
    listCollaborators(request: { path: string }): Promise<{
        collaborators: RepositoryCollaborator[];
    }>;
    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator>;
    removeCollaborator(request: { path: string; userId: number }): Promise<void>;
}
