import { client } from '#electron/bridge.js';
import type { IpcRouter } from '#electron/ipc/router.js';
import type { IpcDeps } from '#electron/ipc/types.js';

export function register(router: IpcRouter, deps: IpcDeps): void {
	router.on('system:stats', async () => (await client()).system.stats());

	router.on('system:openDevTools', () => {
		const mainWindow = deps.getMainWindow();

		if (!mainWindow) {
			return;
		}

		deps.openDevToolsPanel(mainWindow);
	});
}
