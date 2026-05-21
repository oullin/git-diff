import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle("reviews:create", async (_event, request: Record<string, unknown>) =>
        (await client()).reviews.create(request),
    );

    ipcMain.handle("reviews:list", async (_event, limit?: number) =>
        (await client()).reviews.list({ limit }),
    );

    ipcMain.handle("reviews:detail", async (_event, id: string) =>
        (await client()).reviews.detail({ id }),
    );

    ipcMain.handle(
        "reviews:event",
        async (
            _event,
            request: {
                reviewId: string;
                type: string;
                filePath?: string;
                message?: string;
                metadata?: string;
            },
        ) => (await client()).reviews.addEvent(request),
    );

    ipcMain.handle(
        "reviews:comment:create",
        async (
            _event,
            request: {
                reviewId: string;
                filePath: string;
                diffSection: string;
                side: string;
                lineNumber: number;
                authorLabel: string;
                bodyHtml: string;
            },
        ) => (await client()).reviews.createComment(request),
    );

    ipcMain.handle(
        "reviews:comment:update",
        async (
            _event,
            request: {
                reviewId: string;
                commentId: string;
                bodyHtml: string;
            },
        ) => (await client()).reviews.updateComment(request),
    );

    ipcMain.handle(
        "reviews:comment:delete",
        async (
            _event,
            request: {
                reviewId: string;
                commentId: string;
            },
        ) => (await client()).reviews.deleteComment(request),
    );
}
