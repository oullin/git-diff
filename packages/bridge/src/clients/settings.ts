import type { RuntimeSettings, SettingsResponse, UIPreferencesResponse } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function getSettings(socketPath: string): Promise<SettingsResponse> {
  return requestJson<SettingsResponse>(socketPath, "GET", "/v1/settings");
}

export function validateSettings(
  socketPath: string,
  request: { settings: RuntimeSettings },
): Promise<SettingsResponse> {
  return requestJson<SettingsResponse>(socketPath, "POST", "/v1/settings/validate", {
    settings: request.settings,
  });
}

export function getUIPreferences(socketPath: string): Promise<UIPreferencesResponse> {
  return requestJson<UIPreferencesResponse>(socketPath, "GET", "/v1/preferences");
}

export function saveUIPreferences(
  socketPath: string,
  values: Record<string, string>,
): Promise<UIPreferencesResponse> {
  return requestJson<UIPreferencesResponse>(socketPath, "POST", "/v1/preferences", { values });
}
