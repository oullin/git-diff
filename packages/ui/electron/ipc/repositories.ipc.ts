import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
  ipcMain.handle("repositories:list", async () => {
    const response = await (await client()).listRepositories();

    return response.repositories ?? [];
  });

  ipcMain.handle("repositories:search-files", async (_event, query: string, limit?: number) => {
    const response = await (await client()).searchRepositoryFiles({ query, limit });

    return response.results ?? [];
  });

  ipcMain.handle("repositories:upsert", async (_event, request: { path: string; name?: string }) =>
    (await client()).upsertRepository(request),
  );

  ipcMain.handle("repositories:remove", async (_event, path: string) =>
    (await client()).removeRepository({ path }),
  );

  ipcMain.handle("repositories:collaborators:list", async (_event, path: string) => {
    const response = await (await client()).listCollaborators({ path });
    return response.collaborators ?? [];
  });

  ipcMain.handle(
    "repositories:collaborators:add",
    async (_event, request: { path: string; userId: number; role: "write" | "read" }) =>
      (await client()).addCollaborator(request),
  );

  ipcMain.handle(
    "repositories:collaborators:remove",
    async (_event, request: { path: string; userId: number }) =>
      (await client()).removeCollaborator(request),
  );
}
