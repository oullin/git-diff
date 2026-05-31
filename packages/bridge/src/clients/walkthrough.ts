import type { WalkthroughRecord } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class WalkthroughClient {
	constructor(private readonly transport: HttpTransport) {}

	generate(request: { path?: string; kind?: 'working' | 'commit'; sha?: string; refresh?: boolean }): Promise<WalkthroughRecord> {
		return this.transport.request<WalkthroughRecord>('POST', HttpRoutes.walkthrough.generate, {
			path: request.path ?? '',
			kind: request.kind ?? 'working',
			sha: request.sha ?? '',
			refresh: request.refresh ?? false,
		});
	}
}
