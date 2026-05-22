import { BrowserWindow, dialog, type OpenDialogOptions } from "electron";
import { client } from "#electron/bridge.js";
import { createWindow, takeIntentForWindow } from "#electron/windows.js";
import type { IpcRouter } from "#electron/ipc/router.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function register(router: IpcRouter, deps: IpcDeps): void {
    router.on("repository:state", async (_event, path?: string) =>
        (await client()).repository.state({ path }),
    );

    router.on("repository:open", async (_event, path: string) =>
        (await client()).repository.open({ path }),
    );

    router.on("repository:refresh", async (_event, path: string) =>
        (await client()).repository.refresh({ path }),
    );

    router.on("repository:commit", async (_event, sha: string, path?: string) =>
        (await client()).repository.readCommit({ path, sha }),
    );

    router.on("repository:log", async (_event, path?: string, limit?: number) =>
        (await client()).repository.listCommits({ path, limit }),
    );

    router.on("repository:file:read", async (_event, root: string, path: string) =>
        (await client()).repository.readFile({ root, path }),
    );

    router.on(
        "repository:file:range",
        async (
            _event,
            request: {
                root: string;
                path: string;
                ref?: string;
                startLine: number;
                endLine: number;
            },
        ) => (await client()).repository.readFileRange(request),
    );

    router.on("repository:choose", async (_event, defaultPath?: string) => {
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

    router.on("launch-intent:take", async (event) => {
        const window = BrowserWindow.fromWebContents(event.sender);

        return window ? takeIntentForWindow(window) : null;
    });

    router.on("window:new", async (_event, repoPath?: string) => {
        const intent = repoPath ? { kind: "working" as const, repoPath, walkthrough: false } : null;

        createWindow(intent);
    });
}
