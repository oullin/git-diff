import type { FileSearchResult, Repository, RepositoryCollaborator } from "@git-diff/contracts";
export declare function listRepositories(socketPath: string): Promise<{
    repositories: Repository[];
}>;
export declare function searchRepositoryFiles(
    socketPath: string,
    request: {
        query: string;
        limit?: number;
    },
): Promise<{
    results: FileSearchResult[];
}>;
export declare function upsertRepository(
    socketPath: string,
    request: {
        path: string;
        name?: string;
    },
): Promise<Repository>;
export declare function removeRepository(
    socketPath: string,
    request: {
        path: string;
    },
): Promise<void>;
export declare function listCollaborators(
    socketPath: string,
    request: {
        path: string;
    },
): Promise<{
    collaborators: RepositoryCollaborator[];
}>;
export declare function addCollaborator(
    socketPath: string,
    request: {
        path: string;
        userId: number;
        role: "write" | "read";
    },
): Promise<RepositoryCollaborator>;
export declare function removeCollaborator(
    socketPath: string,
    request: {
        path: string;
        userId: number;
    },
): Promise<void>;
