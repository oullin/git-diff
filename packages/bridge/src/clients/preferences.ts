import type { UserPreferences } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';
import { HttpRoutes } from '#bridge/routes.js';

export class PreferenceClient {
	constructor(private readonly transport: HttpTransport) {}

	get(): Promise<UserPreferences> {
		return this.transport.request<UserPreferences>('GET', HttpRoutes.preferences.base);
	}

	save(values: Record<string, string>): Promise<UserPreferences> {
		return this.transport.request<UserPreferences>('POST', HttpRoutes.preferences.base, { values });
	}
}
