import type { RuntimeSettings, SettingsResponse, UIPreferencesResponse } from "@git-diff/contracts";
export declare function getSettings(socketPath: string): Promise<SettingsResponse>;
export declare function validateSettings(socketPath: string, request: {
    settings: RuntimeSettings;
}): Promise<SettingsResponse>;
export declare function getUIPreferences(socketPath: string): Promise<UIPreferencesResponse>;
export declare function saveUIPreferences(socketPath: string, values: Record<string, string>): Promise<UIPreferencesResponse>;
