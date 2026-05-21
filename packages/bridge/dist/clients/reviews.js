import { requestJson } from "#bridge/http.js";
export function createReview(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/reviews", request);
}
export function listReviews(socketPath, request = {}) {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
    return requestJson(socketPath, "GET", `/v1/reviews${query}`);
}
export function reviewDetail(socketPath, request) {
    return requestJson(socketPath, "GET", `/v1/reviews/${encodeURIComponent(request.id)}`);
}
export function addReviewEvent(socketPath, request) {
    return requestJson(
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
export function createReviewComment(socketPath, request) {
    return requestJson(
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
export function updateReviewComment(socketPath, request) {
    return requestJson(
        socketPath,
        "PATCH",
        `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
        { bodyHtml: request.bodyHtml },
    );
}
export function deleteReviewComment(socketPath, request) {
    return requestJson(
        socketPath,
        "DELETE",
        `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
    );
}
