import type { PendingComment } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function listPendingComments(
  socketPath: string,
  request: { path?: string; kind?: "working" | "commit"; sha?: string },
): Promise<{ comments: PendingComment[] }> {
  const parts: string[] = [];
  if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
  if (request.kind) parts.push(`kind=${request.kind}`);
  if (request.sha) parts.push(`sha=${encodeURIComponent(request.sha)}`);
  const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
  return requestJson<{ comments: PendingComment[] }>(
    socketPath,
    "GET",
    `/v1/pending-comments${query}`,
  );
}

export function createPendingComment(
  socketPath: string,
  request: {
    repoRoot: string;
    contextKind: "working" | "commit";
    contextSha?: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
  },
): Promise<PendingComment> {
  return requestJson<PendingComment>(socketPath, "POST", "/v1/pending-comments", request);
}

export function updatePendingComment(
  socketPath: string,
  request: { id: string; bodyHtml: string },
): Promise<PendingComment> {
  return requestJson<PendingComment>(
    socketPath,
    "PATCH",
    `/v1/pending-comments/${encodeURIComponent(request.id)}`,
    { bodyHtml: request.bodyHtml },
  );
}

export function deletePendingComment(
  socketPath: string,
  request: { id: string },
): Promise<void> {
  return requestJson<void>(
    socketPath,
    "DELETE",
    `/v1/pending-comments/${encodeURIComponent(request.id)}`,
  );
}

export function promotePendingComments(
  socketPath: string,
  request: { reviewId: string },
): Promise<{ promoted: number }> {
  return requestJson<{ promoted: number }>(
    socketPath,
    "POST",
    "/v1/pending-comments/promote",
    request,
  );
}
