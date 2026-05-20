import { requestJson } from "#bridge/http.js";
export function getUIPreferences(socketPath) {
    return requestJson(socketPath, "GET", "/v1/preferences");
}
export function saveUIPreferences(socketPath, values) {
    return requestJson(socketPath, "POST", "/v1/preferences", { values });
}
