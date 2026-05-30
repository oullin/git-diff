import type { FileSearchResult, Repository, RepositoryCollaborator } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";

export class RepositoriesClient {
    constructor(private readonly transport: HttpTransport) {}

    list(): Promise<{ repositories: Repository[] }> {
        return this.transport.request<{ repositories: Repository[] }>("GET", "/v1/repositories");
    }

    searchFiles(request: {
        query: string;
        limit?: number;
    }): Promise<{ results: FileSearchResult[] }> {
        const params = new URLSearchParams({ q: request.query });

        if (typeof request.limit === "number") {
            params.set("limit", String(request.limit));
        }

        return this.transport.request<{ results: FileSearchResult[] }>(
            "GET",
            `/v1/repositories/search-files?${params.toString()}`,
        );
    }

    upsert(request: { path: string; name?: string }): Promise<Repository> {
        return this.transport.request<Repository>("POST", "/v1/repositories", {
            path: request.path,
            name: request.name ?? "",
        });
    }

    remove(request: { path: string }): Promise<void> {
        return this.transport.request<void>(
            "DELETE",
            `/v1/repositories?path=${encodeURIComponent(request.path)}`,
        );
    }

    listCollaborators(request: {
        path: string;
    }): Promise<{ collaborators: RepositoryCollaborator[] }> {
        return this.transport.request<{ collaborators: RepositoryCollaborator[] }>(
            "GET",
            `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`,
        );
    }

    addCollaborator(request: {
        path: string;
        userId: number;
        role: "write" | "read";
    }): Promise<RepositoryCollaborator> {
        return this.transport.request<RepositoryCollaborator>(
            "POST",
            "/v1/repositories/collaborators",
            {
                path: request.path,
                userId: request.userId,
                role: request.role,
            },
        );
    }

    removeCollaborator(request: { path: string; userId: number }): Promise<void> {
        const params = new URLSearchParams({
            path: request.path,
            userId: String(request.userId),
        });

        return this.transport.request<void>(
            "DELETE",
            `/v1/repositories/collaborators?${params.toString()}`,
        );
    }
}
