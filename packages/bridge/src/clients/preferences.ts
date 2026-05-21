import type { UIPreferences } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function getUIPreferences(socketPath: string): Promise<UIPreferences> {
    return requestJson<UIPreferences>(socketPath, "GET", "/v1/preferences");
}

export function saveUIPreferences(
    socketPath: string,
    values: Record<string, string>,
): Promise<UIPreferences> {
    return requestJson<UIPreferences>(socketPath, "POST", "/v1/preferences", { values });
}
