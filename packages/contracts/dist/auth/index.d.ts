export interface AuthUser {
    id: number;
    osUsername: string;
    displayName: string;
}
export interface AuthStateResponse {
    osUsername: string;
    needsSetup: boolean;
    isAuthenticated: boolean;
}
export interface AuthLoginResponse {
    token?: string;
    user: AuthUser;
}
export interface AuthBootstrapResponse {
    user: AuthUser | null;
    state: AuthStateResponse;
}
export interface AuthSetupRequest {
    password: string;
}
export interface AuthLoginRequest {
    password: string;
    remember: boolean;
}
export interface AuthResumeRequest {
    token: string;
}
export interface AuthResumeResponse {
    user: AuthUser;
}
export interface AuthWipeRequest {
    osUsername?: string;
}
//# sourceMappingURL=index.d.ts.map
