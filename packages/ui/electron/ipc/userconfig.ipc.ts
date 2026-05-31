import { shell } from 'electron';
import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.userConfig.get, async () => (await client()).userConfig.get());

	// user-config:open intentionally takes no arguments — the path is sourced
	// from the backend, never the renderer, so a compromised renderer cannot
	// coerce the main process into opening an arbitrary file.
	router.on(IpcChannels.userConfig.open, async () => {
		const cfg = await (await client()).userConfig.get();

		if (!cfg.path) {
			throw new Error('user config path is unavailable');
		}

		const err = await shell.openPath(cfg.path);

		if (err) {
			throw new Error(err);
		}
	});
}
