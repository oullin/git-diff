import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function register(deps: IpcDeps): void {
  ipcMain.handle("system:stats", async () => (await client()).getSystemStats());

  ipcMain.handle("system:openDevTools", () => {
    const mainWindow = deps.getMainWindow();

    if (!mainWindow) {
      return;
    }

    deps.openDevToolsPanel(mainWindow);
  });
}
