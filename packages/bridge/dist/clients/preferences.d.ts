import type { UIPreferencesResponse } from "@git-diff/contracts";
export declare function getUIPreferences(socketPath: string): Promise<UIPreferencesResponse>;
export declare function saveUIPreferences(
    socketPath: string,
    values: Record<string, string>,
): Promise<UIPreferencesResponse>;
