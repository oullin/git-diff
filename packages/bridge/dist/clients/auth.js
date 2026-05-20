import { requestJson } from "#bridge/http.js";
export function getAuthState(socketPath) {
    return requestJson(socketPath, "GET", "/v1/auth/state");
}
export function authSetup(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/auth/setup", {
        password: request.password,
    });
}
export function authLogin(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/auth/login", {
        password: request.password,
        remember: request.remember,
    });
}
export function authResume(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/auth/resume", {
        token: request.token,
    });
}
export function authLogout(socketPath) {
    return requestJson(socketPath, "POST", "/v1/auth/logout");
}
export function authWipe(socketPath, request) {
    return requestJson(socketPath, "POST", "/v1/auth/wipe", {
        osUsername: request.osUsername ?? "",
    });
}
