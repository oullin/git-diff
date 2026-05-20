import { requestJson } from "#bridge/http.js";
export function listRepositories(socketPath) {
    return requestJson(socketPath, "GET", "/v1/repositories");
}
export function searchRepositoryFiles(socketPath, request) {
    const params = new URLSearchParams({ q: request.query });
    if (typeof request.limit === "number") {
        params.set("limit", String(request.limit));
    }
    return requestJson(socketPath, "GET", `/v1/repositories/search-files?${params.toString()}`);
}
export function upsertRepository(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repositories", {
        path: request.path,
        name: request.name ?? "",
    });
}
export function removeRepository(socketPath, request) {
    return requestJson(socketPath, "DELETE", `/v1/repositories?path=${encodeURIComponent(request.path)}`);
}
export function listCollaborators(socketPath, request) {
    return requestJson(socketPath, "GET", `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`);
}
export function addCollaborator(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/repositories/collaborators", {
        path: request.path,
        userId: request.userId,
        role: request.role,
    });
}
export function removeCollaborator(socketPath, request) {
    const params = new URLSearchParams({
        path: request.path,
        userId: String(request.userId),
    });
    return requestJson(socketPath, "DELETE", `/v1/repositories/collaborators?${params.toString()}`);
}
