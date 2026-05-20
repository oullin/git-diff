import { requestJson } from "#bridge/http.js";
import { runWorkflowStream } from "#bridge/sse.js";
export function listWorkflows(socketPath) {
    return requestJson(socketPath, "GET", "/v1/workflows");
}
export function listRuns(socketPath, request = {}) {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
    return requestJson(socketPath, "GET", `/v1/runs${query}`);
}
export function runLog(socketPath, request) {
    return requestJson(socketPath, "GET", `/v1/runs/${encodeURIComponent(request.runId)}/log`);
}
export function listTemplateFiles(socketPath) {
    return requestJson(socketPath, "GET", "/v1/template-files");
}
export function readTemplateFile(socketPath, request) {
    return requestJson(socketPath, "GET", `/v1/template-files/content?path=${encodeURIComponent(request.path)}`);
}
export function saveTemplateFile(socketPath, request) {
    return requestJson(socketPath, "PUT", "/v1/template-files/content", {
        path: request.path,
        content: request.content,
    });
}
export function runWorkflow(socketPath, request) {
    return runWorkflowStream(socketPath, request);
}
