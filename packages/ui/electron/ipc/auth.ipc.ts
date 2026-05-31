import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

import type { AuthLoginRequest, AuthSetupRequest, AuthStateResponse, AuthUser, AuthWipeRequest } from '@git-diff/domain';

import { clearSessionToken, readSessionToken, writeSessionToken } from '#electron/ipc/session-token.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.auth.state, async () => (await client()).auth.getState());

	router.on(IpcChannels.auth.setup, async (_event, request: AuthSetupRequest) => {
		const response = await (await client()).auth.setup(request);

		if (response.token) {
			writeSessionToken(response.token);
		}

		return response;
	});

	router.on(IpcChannels.auth.login, async (_event, request: AuthLoginRequest) => {
		const response = await (await client()).auth.login(request);

		if (request.remember && response.token) {
			writeSessionToken(response.token);
		} else {
			clearSessionToken();
		}

		return response;
	});

	router.on(IpcChannels.auth.logout, async () => {
		clearSessionToken();

		await (await client()).auth.logout();
	});

	router.on(IpcChannels.auth.wipe, async (_event, request: AuthWipeRequest = {}) => {
		clearSessionToken();

		await (await client()).auth.wipe({ osUsername: request.osUsername });
	});

	router.on(IpcChannels.auth.bootstrap, async (): Promise<{ user: AuthUser | null; state: AuthStateResponse }> => {
		const token = readSessionToken();

		const c = await client();

		let user: AuthUser | null = null;

		if (token) {
			try {
				const resumed = await c.auth.resume({ token });

				user = resumed.user;
			} catch {
				clearSessionToken();
			}
		}

		const state = await c.auth.getState();

		return { user, state };
	});
}
