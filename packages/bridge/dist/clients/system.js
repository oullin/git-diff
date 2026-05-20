import { requestJson } from "#bridge/http.js";
export function healthz(socketPath) {
    return requestJson(socketPath, "GET", "/v1/healthz");
}
export function getSystemStats(socketPath) {
    return requestJson(socketPath, "GET", "/v1/system/stats");
}
