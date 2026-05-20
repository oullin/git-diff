import type {
  FileSearchResult,
  Repository,
  RepositoryCollaborator,
} from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function listRepositories(socketPath: string): Promise<{ repositories: Repository[] }> {
  return requestJson<{ repositories: Repository[] }>(socketPath, "GET", "/v1/repositories");
}

export function searchRepositoryFiles(
  socketPath: string,
  request: { query: string; limit?: number },
): Promise<{ results: FileSearchResult[] }> {
  const params = new URLSearchParams({ q: request.query });

  if (typeof request.limit === "number") {
    params.set("limit", String(request.limit));
  }

  return requestJson<{ results: FileSearchResult[] }>(
    socketPath,
    "GET",
    `/v1/repositories/search-files?${params.toString()}`,
  );
}

export function upsertRepository(
  socketPath: string,
  request: { path: string; name?: string },
): Promise<Repository> {
  return requestJson<Repository>(socketPath, "POST", "/v1/repositories", {
    path: request.path,
    name: request.name ?? "",
  });
}

export function removeRepository(socketPath: string, request: { path: string }): Promise<void> {
  return requestJson<void>(
    socketPath,
    "DELETE",
    `/v1/repositories?path=${encodeURIComponent(request.path)}`,
  );
}

export function listCollaborators(
  socketPath: string,
  request: { path: string },
): Promise<{ collaborators: RepositoryCollaborator[] }> {
  return requestJson<{ collaborators: RepositoryCollaborator[] }>(
    socketPath,
    "GET",
    `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`,
  );
}

export function addCollaborator(
  socketPath: string,
  request: { path: string; userId: number; role: "write" | "read" },
): Promise<RepositoryCollaborator> {
  return requestJson<RepositoryCollaborator>(
    socketPath,
    "POST",
    "/v1/repositories/collaborators",
    {
      path: request.path,
      userId: request.userId,
      role: request.role,
    },
  );
}

export function removeCollaborator(
  socketPath: string,
  request: { path: string; userId: number },
): Promise<void> {
  const params = new URLSearchParams({
    path: request.path,
    userId: String(request.userId),
  });
  return requestJson<void>(
    socketPath,
    "DELETE",
    `/v1/repositories/collaborators?${params.toString()}`,
  );
}
