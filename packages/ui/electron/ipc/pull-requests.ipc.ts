import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
  ipcMain.handle("pull-requests:list", async (_event, path?: string, limit?: number) =>
    (await client()).listPullRequests({ path, limit }),
  );

  ipcMain.handle("pull-requests:open", async (_event, number: number, path?: string) =>
    (await client()).readPullRequest({ path, number }),
  );
}
