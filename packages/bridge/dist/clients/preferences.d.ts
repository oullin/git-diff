import type { UIPreferences } from "@git-diff/contracts";
export declare function getUIPreferences(socketPath: string): Promise<UIPreferences>;
export declare function saveUIPreferences(
    socketPath: string,
    values: Record<string, string>,
): Promise<UIPreferences>;
