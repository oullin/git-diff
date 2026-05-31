import type { PullRequestSummary, RepositoryState } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';

export class PullRequestClient {
	constructor(private readonly transport: HttpTransport) {}

	list(request: { path?: string; limit?: number }): Promise<{ pullRequests: PullRequestSummary[] }> {
		const parts: string[] = [];

		if (request.path) {
			parts.push(`path=${encodeURIComponent(request.path)}`);
		}

		if (request.limit) {
			parts.push(`limit=${request.limit}`);
		}

		const query = parts.length === 0 ? '' : `?${parts.join('&')}`;

		return this.transport.request<{ pullRequests: PullRequestSummary[] }>('GET', `/v1/repository/pull-requests${query}`);
	}

	read(request: { path?: string; number: number }): Promise<RepositoryState> {
		const parts = [`number=${request.number}`];

		if (request.path) {
			parts.push(`path=${encodeURIComponent(request.path)}`);
		}

		return this.transport.request<RepositoryState>('GET', `/v1/repository/pull-request?${parts.join('&')}`);
	}
}
