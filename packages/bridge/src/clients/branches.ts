import type { Branch, RepositoryState } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';

export class BranchClient {
	constructor(private readonly transport: HttpTransport) {}

	list(request: { path?: string }): Promise<{ branches: string[]; records?: Branch[] }> {
		const query = request.path ? `?path=${encodeURIComponent(request.path)}` : '';

		return this.transport.request<{ branches: string[]; records?: Branch[] }>('GET', `/v1/repository/branches${query}`);
	}

	checkout(request: { path: string; branch: string }): Promise<RepositoryState> {
		return this.transport.request<RepositoryState>('POST', '/v1/repository/checkout', {
			path: request.path,
			branch: request.branch,
		});
	}

	create(request: { path: string; name: string }): Promise<RepositoryState> {
		return this.transport.request<RepositoryState>('POST', '/v1/repository/branches/create', {
			path: request.path,
			name: request.name,
		});
	}

	delete(request: { path?: string; name: string }): Promise<void> {
		const params = new URLSearchParams({ name: request.name });

		if (request.path) {
			params.set('path', request.path);
		}

		return this.transport.request<void>('DELETE', `/v1/repository/branches?${params.toString()}`);
	}

	lock(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
		return this.transport.request<{ branches: Branch[] }>('POST', '/v1/repository/branches/lock', {
			path: request.path,
			name: request.name,
		});
	}

	unlock(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
		return this.transport.request<{ branches: Branch[] }>('POST', '/v1/repository/branches/unlock', {
			path: request.path,
			name: request.name,
		});
	}
}
