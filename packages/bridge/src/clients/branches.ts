import type { Branch, RepositoryState } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function listBranches(
  socketPath: string,
  request: { path?: string },
): Promise<{ branches: string[]; records?: Branch[] }> {
  const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
  return requestJson<{ branches: string[]; records?: Branch[] }>(
    socketPath,
    "GET",
    `/v1/repository/branches${query}`,
  );
}

export function checkoutBranch(
  socketPath: string,
  request: { path: string; branch: string },
): Promise<RepositoryState> {
  return requestJson<RepositoryState>(socketPath, "POST", "/v1/repository/checkout", {
    path: request.path,
    branch: request.branch,
  });
}

export function createBranch(
  socketPath: string,
  request: { path: string; name: string },
): Promise<RepositoryState> {
  return requestJson<RepositoryState>(socketPath, "POST", "/v1/repository/branches/create", {
    path: request.path,
    name: request.name,
  });
}

export function deleteBranch(
  socketPath: string,
  request: { path?: string; name: string },
): Promise<void> {
  const params = new URLSearchParams({ name: request.name });
  if (request.path) params.set("path", request.path);
  return requestJson<void>(socketPath, "DELETE", `/v1/repository/branches?${params.toString()}`);
}

export function lockBranch(
  socketPath: string,
  request: { path?: string; name: string },
): Promise<{ branches: Branch[] }> {
  return requestJson<{ branches: Branch[] }>(socketPath, "POST", "/v1/repository/branches/lock", {
    path: request.path,
    name: request.name,
  });
}

export function unlockBranch(
  socketPath: string,
  request: { path?: string; name: string },
): Promise<{ branches: Branch[] }> {
  return requestJson<{ branches: Branch[] }>(socketPath, "POST", "/v1/repository/branches/unlock", {
    path: request.path,
    name: request.name,
  });
}
