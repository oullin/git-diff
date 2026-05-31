import type { AuthBootstrapResponse, AuthLoginResponse, AuthStateResponse } from '@git-diff/domain';

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
