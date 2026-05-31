import { client } from '#electron/bridge.js';
import { IpcChannels } from '#electron/ipc/routes.js';
import type { IpcRouter } from '#electron/ipc/router.js';

export function register(router: IpcRouter): void {
	router.on(IpcChannels.repositories.list, async () => {
		const response = await (await client()).repositories.list();

		return response.repositories ?? [];
	});

	router.on(IpcChannels.repositories.searchFiles, async (_event, query: string, limit?: number) => {
		const response = await (await client()).repositories.searchFiles({ query, limit });

		return response.results ?? [];
	});

	router.on(IpcChannels.repositories.upsert, async (_event, request: { path: string; name?: string }) => (await client()).repositories.upsert(request));

	router.on(IpcChannels.repositories.remove, async (_event, path: string) => (await client()).repositories.remove({ path }));

	router.on(IpcChannels.repositories.collaboratorsList, async (_event, path: string) => {
		const response = await (await client()).repositories.listCollaborators({ path });

		return response.collaborators ?? [];
	});

	router.on(IpcChannels.repositories.collaboratorsAdd, async (_event, request: { path: string; userId: number; role: 'write' | 'read' }) => (await client()).repositories.addCollaborator(request));

	router.on(IpcChannels.repositories.collaboratorsRemove, async (_event, request: { path: string; userId: number }) => (await client()).repositories.removeCollaborator(request));
}
