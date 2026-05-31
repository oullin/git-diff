import type { PendingComment } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class PendingCommentClient {
	constructor(private readonly transport: HttpTransport) {}

	list(request: { path?: string; kind?: 'working' | 'commit'; sha?: string }): Promise<{ comments: PendingComment[] }> {
		const parts: string[] = [];

		if (request.path) {
			parts.push(`path=${encodeURIComponent(request.path)}`);
		}

		if (request.kind) {
			parts.push(`kind=${request.kind}`);
		}

		if (request.sha) {
			parts.push(`sha=${encodeURIComponent(request.sha)}`);
		}

		const query = parts.length === 0 ? '' : `?${parts.join('&')}`;

		return this.transport.request<{ comments: PendingComment[] }>('GET', `${HttpRoutes.pendingComments.base}${query}`);
	}

	create(request: {
		repoRoot: string;
		contextKind: 'working' | 'commit';
		contextSha?: string;
		filePath: string;
		diffSection: string;
		side: string;
		lineNumber: number;
		startLineNumber?: number;
		startSide?: string;
		authorLabel: string;
		bodyHtml: string;
	}): Promise<PendingComment> {
		return this.transport.request<PendingComment>('POST', HttpRoutes.pendingComments.base, request);
	}

	update(request: { id: number; bodyHtml: string }): Promise<PendingComment> {
		return this.transport.request<PendingComment>('PATCH', HttpRoutes.pendingComments.item(request.id), {
			bodyHtml: request.bodyHtml,
		});
	}

	delete(request: { id: number }): Promise<void> {
		return this.transport.request<void>('DELETE', HttpRoutes.pendingComments.item(request.id));
	}

	promote(request: { reviewId: number }): Promise<{ promoted: number }> {
		return this.transport.request<{ promoted: number }>('POST', HttpRoutes.pendingComments.promote, request);
	}
}
