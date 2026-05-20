import { requestJson } from "#bridge/http.js";
export function listOpVaults(socketPath) {
    return requestJson(socketPath, "GET", "/v1/onepassword/vaults");
}
export function listOpItems(socketPath, request) {
    return requestJson(socketPath, "GET", `/v1/onepassword/items?vault=${encodeURIComponent(request.vault)}`);
}
