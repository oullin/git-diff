import type { UserConfig } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

/** Hot-reload SSE is separate — subscribe to GET /v1/userconfig/stream
 *  with an EventSource directly. */
export class UserConfigClient {
	constructor(private readonly transport: HttpTransport) {}

	get(): Promise<UserConfig> {
		return this.transport.request<UserConfig>('GET', HttpRoutes.userConfig.base);
	}
}
