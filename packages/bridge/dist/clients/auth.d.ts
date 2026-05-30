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
export declare class AuthClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    getState(): Promise<AuthStateResponse>;
    setup(request: AuthSetupRequest): Promise<AuthLoginResponse>;
    login(request: AuthLoginRequest): Promise<AuthLoginResponse>;
    resume(request: AuthResumeRequest): Promise<AuthResumeResponse>;
    logout(): Promise<void>;
    wipe(request: AuthWipeRequest): Promise<void>;
}
