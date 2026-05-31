import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.branches.list, async (_event, path?: string) => (await client()).branches.list({ path }));

	router.on(IpcChannels.branches.checkout, async (_event, path: string, branch: string) => {
		try {
			return await (await client()).branches.checkout({ path, branch });
		} catch (cause) {
			const err = cause as { code?: string; files?: string[]; message?: string };

			if (err && typeof err.code === 'string') {
				const payload = JSON.stringify({
					code: err.code,
					files: Array.isArray(err.files) ? err.files : [],
					message: typeof err.message === 'string' ? err.message : '',
				});

				throw new Error(`__BRIDGE_ERROR__${payload}`);
			}

			throw cause;
		}
	});

	router.on(IpcChannels.branches.create, async (_event, path: string, name: string) => (await client()).branches.create({ path, name }));

	router.on(IpcChannels.branches.delete, async (_event, name: string, path?: string) => (await client()).branches.delete({ path, name }));

	router.on(IpcChannels.branches.lock, async (_event, name: string, path?: string) => (await client()).branches.lock({ path, name }));

	router.on(IpcChannels.branches.unlock, async (_event, name: string, path?: string) => (await client()).branches.unlock({ path, name }));
}
