import type { UserPreferences } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
export declare class PreferenceClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    get(): Promise<UserPreferences>;
    save(values: Record<string, string>): Promise<UserPreferences>;
}
