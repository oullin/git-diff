import { ipcMain } from "electron";
import os from "node:os";
import { client } from "#electron/bridge.js";
import { accountAvatarUrl, architectureLabel, osLabel } from "#electron/system-info.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function register(deps: IpcDeps): void {
  ipcMain.handle("system:stats", async () => (await client()).getSystemStats());

  // TODO(phase-21): rename the system:mac* channels to system:username / system:hostname /
  // system:systemInfo. Keeping the existing names for now so the preload + renderer
  // continue to resolve.
  ipcMain.handle("system:macName", () => os.userInfo().username);
  ipcMain.handle("system:macHostname", () => os.hostname());
  ipcMain.handle("system:macSystemInfo", () => ({
    name: os.userInfo().username,
    hostname: os.hostname(),
    osLabel: osLabel(),
    architectureLabel: architectureLabel(os.arch()),
    avatarUrl: accountAvatarUrl(),
  }));

  ipcMain.handle("system:openDevTools", () => {
    const mainWindow = deps.getMainWindow();

    if (!mainWindow) {
      return;
    }

    deps.openDevToolsPanel(mainWindow);
  });
}
