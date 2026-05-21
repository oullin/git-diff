import type {
    AuthBootstrapResponse,
    AuthLoginResponse,
    AuthStateResponse,
} from "@git-diff/contracts";

/**
 * AuthService wraps the auth surface of `window.diffApp` so the rest of
 * the app depends on a small, typed interface instead of the global IPC
 * shim. Tests can substitute an in-memory implementation by constructing
 * an alternative bridge and passing it to the service registry.
 */
export interface AuthService {
    state(): Promise<AuthStateResponse>;
    bootstrap(): Promise<AuthBootstrapResponse>;
    setup(password: string): Promise<AuthLoginResponse>;
    login(password: string, remember: boolean): Promise<AuthLoginResponse>;
    logout(): Promise<void>;
    wipe(osUsername?: string): Promise<void>;
}

export function createAuthService(): AuthService {
    return {
        state: () => window.diffApp.getAuthState(),
        bootstrap: () => window.diffApp.authBootstrap(),
        setup: (password) => window.diffApp.authSetup(password),
        login: (password, remember) => window.diffApp.authLogin(password, remember),
        logout: () => window.diffApp.authLogout(),
        wipe: (osUsername) => window.diffApp.authWipe(osUsername),
    };
}
