import type { UIPreferences } from "@git-diff/contracts";

export interface PreferenceService {
    get(): Promise<UIPreferences>;
    save(patch: Record<string, string>): Promise<UIPreferences>;
}

export function createPreferenceService(): PreferenceService {
    return {
        get: () => window.diffApp.getUIPreferences(),
        save: (patch) => window.diffApp.saveUIPreferences(patch),
    };
}
