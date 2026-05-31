import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.preferences.get, async () => (await client()).preferences.get());

	router.on(IpcChannels.preferences.save, async (_event, patch: Record<string, string>) => (await client()).preferences.save(patch ?? {}));
}
