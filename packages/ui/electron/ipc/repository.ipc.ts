import { BrowserWindow, dialog, ipcMain, type OpenDialogOptions } from "electron";
import { client } from "#electron/bridge.js";
import { createWindow, takeIntentForWindow } from "#electron/windows.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function register(deps: IpcDeps): void {
    ipcMain.handle("repository:state", async (_event, path?: string) =>
        (await client()).repositoryState({ path }),
    );

    ipcMain.handle("repository:open", async (_event, path: string) =>
        (await client()).openRepository({ path }),
    );

    ipcMain.handle("repository:refresh", async (_event, path: string) =>
        (await client()).refreshRepository({ path }),
    );

    ipcMain.handle("repository:commit", async (_event, sha: string, path?: string) =>
        (await client()).readCommit({ path, sha }),
    );

    ipcMain.handle("repository:log", async (_event, path?: string, limit?: number) =>
        (await client()).listCommits({ path, limit }),
    );

    ipcMain.handle("repository:file:read", async (_event, root: string, path: string) =>
        (await client()).readRepositoryFile({ root, path }),
    );

    ipcMain.handle("repository:choose", async (_event, defaultPath?: string) => {
        const options: OpenDialogOptions = {
            defaultPath,
            properties: ["openDirectory"],
        };
        const mainWindow = deps.getMainWindow();
        const result = mainWindow
            ? await dialog.showOpenDialog(mainWindow, options)
            : await dialog.showOpenDialog(options);

        return result.canceled ? null : (result.filePaths[0] ?? null);
    });

    ipcMain.handle("launch-intent:take", async (event) => {
        const window = BrowserWindow.fromWebContents(event.sender);
        return window ? takeIntentForWindow(window) : null;
    });

    ipcMain.handle("window:new", async (_event, repoPath?: string) => {
        const intent = repoPath ? { kind: "working" as const, repoPath, walkthrough: false } : null;
        createWindow(intent);
    });
}
