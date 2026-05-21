import { requestJson } from "#bridge/http.js";
export function listBranches(socketPath, request) {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return requestJson(socketPath, "GET", `/v1/repository/branches${query}`);
}
export function checkoutBranch(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/checkout", {
        path: request.path,
        branch: request.branch,
    });
}
export function createBranch(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/branches/create", {
        path: request.path,
        name: request.name,
    });
}
export function deleteBranch(socketPath, request) {
    const params = new URLSearchParams({ name: request.name });
    if (request.path) params.set("path", request.path);
    return requestJson(socketPath, "DELETE", `/v1/repository/branches?${params.toString()}`);
}
export function lockBranch(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/branches/lock", {
        path: request.path,
        name: request.name,
    });
}
export function unlockBranch(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repository/branches/unlock", {
        path: request.path,
        name: request.name,
    });
}
