import { ipcMain } from 'electron';
import { register as registerAuth } from '#electron/ipc/auth.ipc.js';
import { register as registerBranches } from '#electron/ipc/branches.ipc.js';
import { register as registerPendingComments } from '#electron/ipc/pending-comments.ipc.js';
import { register as registerPreferences } from '#electron/ipc/preferences.ipc.js';
import { register as registerPullRequests } from '#electron/ipc/pull-requests.ipc.js';
import { register as registerRepositories } from '#electron/ipc/repositories.ipc.js';
import { register as registerRepository } from '#electron/ipc/repository.ipc.js';
import { register as registerReviews } from '#electron/ipc/reviews.ipc.js';
import { IpcRouter } from '#electron/ipc/router.js';
import { register as registerSystem } from '#electron/ipc/system.ipc.js';
import type { IpcDeps } from '#electron/ipc/types.js';
import { register as registerUserConfig } from '#electron/ipc/userconfig.ipc.js';
import { register as registerWalkthrough } from '#electron/ipc/walkthrough.ipc.js';

export function registerIpcHandlers(deps: IpcDeps): void {
	const router = buildRouter(deps);

	router.register(ipcMain);
}

export function buildRouter(deps: IpcDeps): IpcRouter {
	const router = new IpcRouter();

	registerRepository(router, deps);
	registerRepositories(router);
	registerBranches(router);
	registerPullRequests(router);
	registerWalkthrough(router);
	registerPendingComments(router);
	registerPreferences(router);
	registerReviews(router);
	registerAuth(router);
	registerSystem(router, deps);
	registerUserConfig(router);

	return router;
}
