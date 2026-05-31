import { client } from '#electron/bridge.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(
		'walkthrough:generate',
		async (
			_event,
			request: {
				path?: string;
				kind?: 'working' | 'commit';
				sha?: string;
				refresh?: boolean;
			},
		) => (await client()).walkthroughs.generate(request),
	);
}
