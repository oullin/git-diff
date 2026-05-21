import type { AuthLoginResponse, AuthStateResponse, AuthUser } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function getAuthState(socketPath: string): Promise<AuthStateResponse> {
    return requestJson<AuthStateResponse>(socketPath, "GET", "/v1/auth/state");
}

export function authSetup(
    socketPath: string,
    request: { password: string },
): Promise<AuthLoginResponse> {
    return requestJson<AuthLoginResponse>(socketPath, "POST", "/v1/auth/setup", {
        password: request.password,
    });
}

export function authLogin(
    socketPath: string,
    request: { password: string; remember: boolean },
): Promise<AuthLoginResponse> {
    return requestJson<AuthLoginResponse>(socketPath, "POST", "/v1/auth/login", {
        password: request.password,
        remember: request.remember,
    });
}

export function authResume(
    socketPath: string,
    request: { token: string },
): Promise<{ user: AuthUser }> {
    return requestJson<{ user: AuthUser }>(socketPath, "POST", "/v1/auth/resume", {
        token: request.token,
    });
}

export function authLogout(socketPath: string): Promise<void> {
    return requestJson<void>(socketPath, "POST", "/v1/auth/logout");
}

export function authWipe(socketPath: string, request: { osUsername?: string }): Promise<void> {
    return requestJson<void>(socketPath, "POST", "/v1/auth/wipe", {
        osUsername: request.osUsername ?? "",
    });
}
