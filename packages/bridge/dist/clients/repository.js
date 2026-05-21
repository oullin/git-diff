import { requestJson } from "#bridge/http.js";
export function repositoryState(socketPath, request) {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return requestJson(socketPath, "GET", `/v1/repository/state${query}`);
}
export function openRepository(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/open", {
        path: request.path,
    });
}
export function refreshRepository(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/refresh", {
        path: request.path,
    });
}
export function readCommit(socketPath, request) {
    const parts = [`sha=${encodeURIComponent(request.sha)}`];
    if (request.path) {
        parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    return requestJson(socketPath, "GET", `/v1/repository/commit?${parts.join("&")}`);
}
export function listCommits(socketPath, request) {
    const parts = [];
    if (request.path) {
        parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    if (request.limit) {
        parts.push(`limit=${request.limit}`);
    }
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return requestJson(socketPath, "GET", `/v1/repository/log${query}`);
}
export function readRepositoryFile(socketPath, request) {
    const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;
    return requestJson(socketPath, "GET", `/v1/repository/file${query}`);
}
