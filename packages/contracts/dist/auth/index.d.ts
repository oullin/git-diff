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
//# sourceMappingURL=index.d.ts.map
