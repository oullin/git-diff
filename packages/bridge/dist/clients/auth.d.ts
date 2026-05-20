import type { AuthLoginResponse, AuthStateResponse, AuthUser } from "@git-diff/contracts";
export declare function getAuthState(socketPath: string): Promise<AuthStateResponse>;
export declare function authSetup(socketPath: string, request: {
    password: string;
}): Promise<AuthLoginResponse>;
export declare function authLogin(socketPath: string, request: {
    password: string;
    remember: boolean;
}): Promise<AuthLoginResponse>;
export declare function authResume(socketPath: string, request: {
    token: string;
}): Promise<{
    user: AuthUser;
}>;
export declare function authLogout(socketPath: string): Promise<void>;
export declare function authWipe(socketPath: string, request: {
    osUsername?: string;
}): Promise<void>;
