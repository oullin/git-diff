import type { SystemStats } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class SystemClient {
	constructor(private readonly transport: HttpTransport) {}

	healthz(): Promise<void> {
		return this.transport.request<void>('GET', HttpRoutes.system.healthz);
	}

	stats(): Promise<SystemStats> {
		return this.transport.request<SystemStats>('GET', HttpRoutes.system.stats);
	}
}
