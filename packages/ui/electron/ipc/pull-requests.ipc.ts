import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.pullRequests.list, async (_event, path?: string, limit?: number) => (await client()).pullRequests.list({ path, limit }));

	router.on(IpcChannels.pullRequests.open, async (_event, number: number, path?: string) => (await client()).pullRequests.read({ path, number }));
}
