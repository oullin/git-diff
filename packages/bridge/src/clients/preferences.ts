import type { UIPreferencesResponse } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function getUIPreferences(socketPath: string): Promise<UIPreferencesResponse> {
    return requestJson<UIPreferencesResponse>(socketPath, "GET", "/v1/preferences");
}

export function saveUIPreferences(
    socketPath: string,
    values: Record<string, string>,
): Promise<UIPreferencesResponse> {
    return requestJson<UIPreferencesResponse>(socketPath, "POST", "/v1/preferences", { values });
}
