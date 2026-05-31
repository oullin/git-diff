import type { FileSearchResult, Repository, RepositoryCollaborator } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class RepositoriesClient {
	constructor(private readonly transport: HttpTransport) {}

	list(): Promise<{ repositories: Repository[] }> {
		return this.transport.request<{ repositories: Repository[] }>('GET', HttpRoutes.repositories.base);
	}

	searchFiles(request: { query: string; limit?: number }): Promise<{ results: FileSearchResult[] }> {
		const params = new URLSearchParams({ q: request.query });

		if (typeof request.limit === 'number') {
			params.set('limit', String(request.limit));
		}

		return this.transport.request<{ results: FileSearchResult[] }>('GET', `${HttpRoutes.repositories.searchFiles}?${params.toString()}`);
	}

	upsert(request: { path: string; name?: string }): Promise<Repository> {
		return this.transport.request<Repository>('POST', HttpRoutes.repositories.base, {
			path: request.path,
			name: request.name ?? '',
		});
	}

	remove(request: { path: string }): Promise<void> {
		return this.transport.request<void>('DELETE', `${HttpRoutes.repositories.base}?path=${encodeURIComponent(request.path)}`);
	}

	listCollaborators(request: { path: string }): Promise<{ collaborators: RepositoryCollaborator[] }> {
		return this.transport.request<{ collaborators: RepositoryCollaborator[] }>('GET', `${HttpRoutes.repositories.collaborators}?path=${encodeURIComponent(request.path)}`);
	}

	addCollaborator(request: { path: string; userId: number; role: 'write' | 'read' }): Promise<RepositoryCollaborator> {
		return this.transport.request<RepositoryCollaborator>('POST', HttpRoutes.repositories.collaborators, {
			path: request.path,
			userId: request.userId,
			role: request.role,
		});
	}

	removeCollaborator(request: { path: string; userId: number }): Promise<void> {
		const params = new URLSearchParams({
			path: request.path,
			userId: String(request.userId),
		});

		return this.transport.request<void>('DELETE', `${HttpRoutes.repositories.collaborators}?${params.toString()}`);
	}
}
