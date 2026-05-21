import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle("ui-prefs:get", async () => (await client()).getUIPreferences());

    ipcMain.handle("ui-prefs:save", async (_event, patch: Record<string, string>) =>
        (await client()).saveUIPreferences(patch ?? {}),
    );
}
