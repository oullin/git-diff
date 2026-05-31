import { client } from '#electron/bridge.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on('pending-comments:list', async (_event, request: { path?: string; kind?: 'working' | 'commit'; sha?: string }) => (await client()).pendingComments.list(request));

	router.on('pending-comments:create', async (_event, request: Parameters<Awaited<ReturnType<typeof client>>['pendingComments']['create']>[0]) => (await client()).pendingComments.create(request));

	router.on('pending-comments:update', async (_event, request: { id: number; bodyHtml: string }) => (await client()).pendingComments.update(request));

	router.on('pending-comments:delete', async (_event, id: number) => (await client()).pendingComments.delete({ id }));

	router.on('pending-comments:promote', async (_event, reviewId: number) => (await client()).pendingComments.promote({ reviewId }));
}
