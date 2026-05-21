import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle(
        "pending-comments:list",
        async (_event, request: { path?: string; kind?: "working" | "commit"; sha?: string }) =>
            (await client()).pendingComments.list(request),
    );

    ipcMain.handle(
        "pending-comments:create",
        async (
            _event,
            request: Parameters<Awaited<ReturnType<typeof client>>["pendingComments"]["create"]>[0],
        ) => (await client()).pendingComments.create(request),
    );

    ipcMain.handle(
        "pending-comments:update",
        async (_event, request: { id: string; bodyHtml: string }) =>
            (await client()).pendingComments.update(request),
    );

    ipcMain.handle("pending-comments:delete", async (_event, id: string) =>
        (await client()).pendingComments.delete({ id }),
    );

    ipcMain.handle("pending-comments:promote", async (_event, reviewId: string) =>
        (await client()).pendingComments.promote({ reviewId }),
    );
}
