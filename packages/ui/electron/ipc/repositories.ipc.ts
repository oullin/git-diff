import { client } from "#electron/bridge.js";
import type { IpcRouter } from "#electron/ipc/router.js";

export function register(router: IpcRouter): void {
  router.on("repositories:list", async () => {
    const response = await (await client()).repositories.list();

    return response.repositories ?? [];
  });

  router.on("repositories:search-files", async (_event, query: string, limit?: number) => {
    const response = await (await client()).repositories.searchFiles({ query, limit });

    return response.results ?? [];
  });

  router.on("repositories:upsert", async (_event, request: { path: string; name?: string }) =>
    (await client()).repositories.upsert(request),
  );

  router.on("repositories:remove", async (_event, path: string) =>
    (await client()).repositories.remove({ path }),
  );

  router.on("repositories:collaborators:list", async (_event, path: string) => {
    const response = await (await client()).repositories.listCollaborators({ path });

    return response.collaborators ?? [];
  });

  router.on(
    "repositories:collaborators:add",
    async (_event, request: { path: string; userId: number; role: "write" | "read" }) =>
      (await client()).repositories.addCollaborator(request),
  );

  router.on(
    "repositories:collaborators:remove",
    async (_event, request: { path: string; userId: number }) =>
      (await client()).repositories.removeCollaborator(request),
  );
}
