import { requestJson } from "#bridge/http.js";
export function listPendingComments(socketPath, request) {
    const parts = [];
    if (request.path)
        parts.push(`path=${encodeURIComponent(request.path)}`);
    if (request.kind)
        parts.push(`kind=${request.kind}`);
    if (request.sha)
        parts.push(`sha=${encodeURIComponent(request.sha)}`);
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return requestJson(socketPath, "GET", `/v1/pending-comments${query}`);
}
export function createPendingComment(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/pending-comments", request);
}
export function updatePendingComment(socketPath, request) {
    return requestJson(socketPath, "PATCH", `/v1/pending-comments/${encodeURIComponent(request.id)}`, { bodyHtml: request.bodyHtml });
}
export function deletePendingComment(socketPath, request) {
    return requestJson(socketPath, "DELETE", `/v1/pending-comments/${encodeURIComponent(request.id)}`);
}
export function promotePendingComments(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/pending-comments/promote", request);
}
