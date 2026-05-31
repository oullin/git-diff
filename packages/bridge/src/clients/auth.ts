import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

import type { AuthLoginRequest, AuthLoginResponse, AuthResumeRequest, AuthResumeResponse, AuthSetupRequest, AuthStateResponse, AuthWipeRequest } from '@git-diff/domain';

export class AuthClient {
	constructor(private readonly transport: HttpTransport) {}

	getState(): Promise<AuthStateResponse> {
		return this.transport.request<AuthStateResponse>('GET', HttpRoutes.auth.state);
	}

	setup(request: AuthSetupRequest): Promise<AuthLoginResponse> {
		return this.transport.request<AuthLoginResponse>('POST', HttpRoutes.auth.setup, {
			password: request.password,
		});
	}

	login(request: AuthLoginRequest): Promise<AuthLoginResponse> {
		return this.transport.request<AuthLoginResponse>('POST', HttpRoutes.auth.login, {
			password: request.password,
			remember: request.remember,
		});
	}

	resume(request: AuthResumeRequest): Promise<AuthResumeResponse> {
		return this.transport.request<AuthResumeResponse>('POST', HttpRoutes.auth.resume, {
			token: request.token,
		});
	}

	logout(): Promise<void> {
		return this.transport.request<void>('POST', HttpRoutes.auth.logout);
	}

	wipe(request: AuthWipeRequest): Promise<void> {
		return this.transport.request<void>('POST', HttpRoutes.auth.wipe, {
			osUsername: request.osUsername ?? '',
		});
	}
}
