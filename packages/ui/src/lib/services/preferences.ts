import type { UserPreferences } from "@git-diff/domain";

export interface PreferenceService {
    get(): Promise<UserPreferences>;
    save(patch: Record<string, string>): Promise<UserPreferences>;
}

export function createPreferenceService(): PreferenceService {
    return {
        get: () => window.diffApp.getUIPreferences(),
        save: (patch) => window.diffApp.saveUIPreferences(patch),
    };
}
