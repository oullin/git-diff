import type { CommitSummary, RepositoryFile, RepositoryState } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function repositoryState(
  socketPath: string,
  request: { path?: string },
): Promise<RepositoryState> {
  const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
  return requestJson<RepositoryState>(socketPath, "GET", `/v1/repository/state${query}`);
}

export function openRepository(
  socketPath: string,
  request: { path: string },
): Promise<RepositoryState> {
  return requestJson<RepositoryState>(socketPath, "POST", "/v1/repository/open", {
    path: request.path,
  });
}

export function refreshRepository(
  socketPath: string,
  request: { path: string },
): Promise<RepositoryState> {
  return requestJson<RepositoryState>(socketPath, "POST", "/v1/repository/refresh", {
    path: request.path,
  });
}

export function readCommit(
  socketPath: string,
  request: { path?: string; sha: string },
): Promise<RepositoryState> {
  const parts = [`sha=${encodeURIComponent(request.sha)}`];
  if (request.path) {
    parts.push(`path=${encodeURIComponent(request.path)}`);
  }
  return requestJson<RepositoryState>(
    socketPath,
    "GET",
    `/v1/repository/commit?${parts.join("&")}`,
  );
}

export function listCommits(
  socketPath: string,
  request: { path?: string; limit?: number },
): Promise<{ commits: CommitSummary[] }> {
  const parts: string[] = [];
  if (request.path) {
    parts.push(`path=${encodeURIComponent(request.path)}`);
  }
  if (request.limit) {
    parts.push(`limit=${request.limit}`);
  }
  const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
  return requestJson<{ commits: CommitSummary[] }>(
    socketPath,
    "GET",
    `/v1/repository/log${query}`,
  );
}

export function readRepositoryFile(
  socketPath: string,
  request: { root: string; path: string },
): Promise<RepositoryFile> {
  const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;
  return requestJson<RepositoryFile>(socketPath, "GET", `/v1/repository/file${query}`);
}
