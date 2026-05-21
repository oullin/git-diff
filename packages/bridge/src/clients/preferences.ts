import type { UIPreferences } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

export class PreferenceClient {
    constructor(private readonly transport: HttpTransport) {}

    get(): Promise<UIPreferences> {
        return this.transport.request<UIPreferences>("GET", "/v1/preferences");
    }

    save(values: Record<string, string>): Promise<UIPreferences> {
        return this.transport.request<UIPreferences>("POST", "/v1/preferences", { values });
    }
}
