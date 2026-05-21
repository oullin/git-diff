import type { UIPreferences } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class PreferenceClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    get(): Promise<UIPreferences>;
    save(values: Record<string, string>): Promise<UIPreferences>;
}
