import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function createReview(
  socketPath: string,
  request: Partial<ReviewSession>,
): Promise<ReviewSession> {
  return requestJson<ReviewSession>(
    socketPath,
    "POST",
    "/v1/reviews",
    request as Record<string, unknown>,
  );
}

export function listReviews(
  socketPath: string,
  request: { limit?: number } = {},
): Promise<{ reviews: ReviewSession[] }> {
  const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
  return requestJson<{ reviews: ReviewSession[] }>(socketPath, "GET", `/v1/reviews${query}`);
}

export function reviewDetail(socketPath: string, request: { id: string }): Promise<ReviewDetail> {
  return requestJson<ReviewDetail>(
    socketPath,
    "GET",
    `/v1/reviews/${encodeURIComponent(request.id)}`,
  );
}

export function addReviewEvent(
  socketPath: string,
  request: {
    reviewId: string;
    type: string;
    filePath?: string;
    message?: string;
    metadata?: string;
  },
): Promise<ReviewEvent> {
  return requestJson<ReviewEvent>(
    socketPath,
    "POST",
    `/v1/reviews/${encodeURIComponent(request.reviewId)}/events`,
    {
      type: request.type,
      filePath: request.filePath,
      message: request.message,
      metadata: request.metadata,
    },
  );
}

export function createReviewComment(
  socketPath: string,
  request: {
    reviewId: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
  },
): Promise<ReviewComment> {
  return requestJson<ReviewComment>(
    socketPath,
    "POST",
    `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments`,
    {
      filePath: request.filePath,
      diffSection: request.diffSection,
      side: request.side,
      lineNumber: request.lineNumber,
      authorLabel: request.authorLabel,
      bodyHtml: request.bodyHtml,
    },
  );
}

export function updateReviewComment(
  socketPath: string,
  request: { reviewId: string; commentId: string; bodyHtml: string },
): Promise<ReviewComment> {
  return requestJson<ReviewComment>(
    socketPath,
    "PATCH",
    `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
    { bodyHtml: request.bodyHtml },
  );
}

export function deleteReviewComment(
  socketPath: string,
  request: { reviewId: string; commentId: string },
): Promise<void> {
  return requestJson<void>(
    socketPath,
    "DELETE",
    `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
  );
}
