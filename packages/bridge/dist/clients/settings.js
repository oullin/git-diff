import { requestJson } from "#bridge/http.js";
export function getSettings(socketPath) {
    return requestJson(socketPath, "GET", "/v1/settings");
}
export function validateSettings(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/settings/validate", {
        settings: request.settings,
    });
}
export function getUIPreferences(socketPath) {
    return requestJson(socketPath, "GET", "/v1/preferences");
}
export function saveUIPreferences(socketPath, values) {
    return requestJson(socketPath, "POST", "/v1/preferences", { values });
}
