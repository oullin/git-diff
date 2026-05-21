import type {
    AuthLoginRequest,
    AuthLoginResponse,
    AuthResumeRequest,
    AuthResumeResponse,
    AuthSetupRequest,
    AuthStateResponse,
    AuthWipeRequest,
} from "@git-diff/contracts";
export declare function getAuthState(socketPath: string): Promise<AuthStateResponse>;
export declare function authSetup(
    socketPath: string,
    request: AuthSetupRequest,
): Promise<AuthLoginResponse>;
export declare function authLogin(
    socketPath: string,
    request: AuthLoginRequest,
): Promise<AuthLoginResponse>;
export declare function authResume(
    socketPath: string,
    request: AuthResumeRequest,
): Promise<AuthResumeResponse>;
export declare function authLogout(socketPath: string): Promise<void>;
export declare function authWipe(socketPath: string, request: AuthWipeRequest): Promise<void>;
