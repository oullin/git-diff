import type { PullRequestSummary, RepositoryState } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function listPullRequests(
  socketPath: string,
  request: { path?: string; limit?: number },
): Promise<{ pullRequests: PullRequestSummary[] }> {
  const parts: string[] = [];
  if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
  if (request.limit) parts.push(`limit=${request.limit}`);
  const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
  return requestJson<{ pullRequests: PullRequestSummary[] }>(
    socketPath,
    "GET",
    `/v1/repository/pull-requests${query}`,
  );
}

export function readPullRequest(
  socketPath: string,
  request: { path?: string; number: number },
): Promise<RepositoryState> {
  const parts = [`number=${request.number}`];
  if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
  return requestJson<RepositoryState>(
    socketPath,
    "GET",
    `/v1/repository/pull-request?${parts.join("&")}`,
  );
}
