import { requestJson } from "#bridge/http.js";
export function listPullRequests(socketPath, request) {
    const parts = [];
    if (request.path)
        parts.push(`path=${encodeURIComponent(request.path)}`);
    if (request.limit)
        parts.push(`limit=${request.limit}`);
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return requestJson(socketPath, "GET", `/v1/repository/pull-requests${query}`);
}
export function readPullRequest(socketPath, request) {
    const parts = [`number=${request.number}`];
    if (request.path)
        parts.push(`path=${encodeURIComponent(request.path)}`);
    return requestJson(socketPath, "GET", `/v1/repository/pull-request?${parts.join("&")}`);
}
