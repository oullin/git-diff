import type { UserPreferences } from '@git-diff/domain';
import type { HttpTransport } from '#bridge/http.js';

export class PreferenceClient {
	constructor(private readonly transport: HttpTransport) {}

	get(): Promise<UserPreferences> {
		return this.transport.request<UserPreferences>('GET', '/v1/preferences');
	}

	save(values: Record<string, string>): Promise<UserPreferences> {
		return this.transport.request<UserPreferences>('POST', '/v1/preferences', { values });
	}
}
