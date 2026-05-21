import type {
    AuthLoginRequest,
    AuthLoginResponse,
    AuthResumeRequest,
    AuthResumeResponse,
    AuthSetupRequest,
    AuthStateResponse,
    AuthWipeRequest,
} from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

export class AuthClient {
    constructor(private readonly transport: HttpTransport) {}

    getState(): Promise<AuthStateResponse> {
        return this.transport.request<AuthStateResponse>("GET", "/v1/auth/state");
    }

    setup(request: AuthSetupRequest): Promise<AuthLoginResponse> {
        return this.transport.request<AuthLoginResponse>("POST", "/v1/auth/setup", {
            password: request.password,
        });
    }

    login(request: AuthLoginRequest): Promise<AuthLoginResponse> {
        return this.transport.request<AuthLoginResponse>("POST", "/v1/auth/login", {
            password: request.password,
            remember: request.remember,
        });
    }

    resume(request: AuthResumeRequest): Promise<AuthResumeResponse> {
        return this.transport.request<AuthResumeResponse>("POST", "/v1/auth/resume", {
            token: request.token,
        });
    }

    logout(): Promise<void> {
        return this.transport.request<void>("POST", "/v1/auth/logout");
    }

    wipe(request: AuthWipeRequest): Promise<void> {
        return this.transport.request<void>("POST", "/v1/auth/wipe", {
            osUsername: request.osUsername ?? "",
        });
    }
}
