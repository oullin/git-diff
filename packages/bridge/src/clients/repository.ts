import type { BytesResponse, HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

import type { CommitSummary, RepositoryFile, RepositoryFileRange, RepositoryState } from '@git-diff/domain';

export class RepositoryClient {
	constructor(private readonly transport: HttpTransport) {}

	state(request: { path?: string }): Promise<RepositoryState> {
		const query = request.path ? `?path=${encodeURIComponent(request.path)}` : '';

		return this.transport.request<RepositoryState>('GET', `${HttpRoutes.repository.state}${query}`);
	}

	open(request: { path: string }): Promise<RepositoryState> {
		return this.transport.request<RepositoryState>('POST', HttpRoutes.repository.open, {
			path: request.path,
		});
	}

	refresh(request: { path: string }): Promise<RepositoryState> {
		return this.transport.request<RepositoryState>('POST', HttpRoutes.repository.refresh, {
			path: request.path,
		});
	}

	readCommit(request: { path?: string; sha: string }): Promise<RepositoryState> {
		const parts = [`sha=${encodeURIComponent(request.sha)}`];

		if (request.path) {
			parts.push(`path=${encodeURIComponent(request.path)}`);
		}

		return this.transport.request<RepositoryState>('GET', `${HttpRoutes.repository.commit}?${parts.join('&')}`);
	}

	listCommits(request: { path?: string; limit?: number }): Promise<{ commits: CommitSummary[] }> {
		const parts: string[] = [];

		if (request.path) {
			parts.push(`path=${encodeURIComponent(request.path)}`);
		}

		if (request.limit) {
			parts.push(`limit=${request.limit}`);
		}

		const query = parts.length === 0 ? '' : `?${parts.join('&')}`;

		return this.transport.request<{ commits: CommitSummary[] }>('GET', `${HttpRoutes.repository.log}${query}`);
	}

	readFile(request: { root: string; path: string }): Promise<RepositoryFile> {
		const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;

		return this.transport.request<RepositoryFile>('GET', `${HttpRoutes.repository.file}${query}`);
	}

	/**
	 * Fetch the raw bytes of a file at a given ref. Empty ref reads the
	 * working tree, ":0" reads the index, otherwise treated as a git ref.
	 * Returns the bytes plus the Content-Type so the renderer can build
	 * an object URL for inline image display.
	 */
	readFileBytes(request: { root?: string; path: string; ref?: string }): Promise<BytesResponse> {
		const parts = [`path=${encodeURIComponent(request.path)}`];

		if (request.root) {
			parts.push(`root=${encodeURIComponent(request.root)}`);
		}

		if (request.ref !== undefined) {
			parts.push(`ref=${encodeURIComponent(request.ref)}`);
		}

		return this.transport.requestBytes(`${HttpRoutes.repository.fileRaw}?${parts.join('&')}`);
	}

	readFileRange(request: { root: string; path: string; ref?: string; startLine: number; endLine: number }): Promise<RepositoryFileRange> {
		const parts = [`root=${encodeURIComponent(request.root)}`, `path=${encodeURIComponent(request.path)}`, `startLine=${request.startLine}`, `endLine=${request.endLine}`];

		if (request.ref) {
			parts.push(`ref=${encodeURIComponent(request.ref)}`);
		}

		return this.transport.request<RepositoryFileRange>('GET', `${HttpRoutes.repository.fileRange}?${parts.join('&')}`);
	}
}
