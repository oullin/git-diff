import { requestJson } from "#bridge/http.js";
export function generateWalkthrough(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/walkthrough", {
        path: request.path ?? "",
        kind: request.kind ?? "working",
        sha: request.sha ?? "",
        refresh: request.refresh ?? false,
    });
}
