import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle("pull-requests:list", async (_event, path?: string, limit?: number) =>
        (await client()).pullRequests.list({ path, limit }),
    );

    ipcMain.handle("pull-requests:open", async (_event, number: number, path?: string) =>
        (await client()).pullRequests.read({ path, number }),
    );
}
