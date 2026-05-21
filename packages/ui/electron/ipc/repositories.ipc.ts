import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle("repositories:list", async () => {
        const response = await (await client()).repositories.list();

        return response.repositories ?? [];
    });

    ipcMain.handle("repositories:search-files", async (_event, query: string, limit?: number) => {
        const response = await (await client()).repositories.searchFiles({ query, limit });

        return response.results ?? [];
    });

    ipcMain.handle(
        "repositories:upsert",
        async (_event, request: { path: string; name?: string }) =>
            (await client()).repositories.upsert(request),
    );

    ipcMain.handle("repositories:remove", async (_event, path: string) =>
        (await client()).repositories.remove({ path }),
    );

    ipcMain.handle("repositories:collaborators:list", async (_event, path: string) => {
        const response = await (await client()).repositories.listCollaborators({ path });
        return response.collaborators ?? [];
    });

    ipcMain.handle(
        "repositories:collaborators:add",
        async (_event, request: { path: string; userId: number; role: "write" | "read" }) =>
            (await client()).repositories.addCollaborator(request),
    );

    ipcMain.handle(
        "repositories:collaborators:remove",
        async (_event, request: { path: string; userId: number }) =>
            (await client()).repositories.removeCollaborator(request),
    );
}
