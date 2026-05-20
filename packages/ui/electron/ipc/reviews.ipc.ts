import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
  ipcMain.handle("reviews:create", async (_event, request: Record<string, unknown>) =>
    (await client()).createReview(request),
  );

  ipcMain.handle("reviews:list", async (_event, limit?: number) =>
    (await client()).listReviews({ limit }),
  );

  ipcMain.handle("reviews:detail", async (_event, id: string) =>
    (await client()).reviewDetail({ id }),
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
    ) => (await client()).addReviewEvent(request),
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
    ) => (await client()).createReviewComment(request),
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
    ) => (await client()).updateReviewComment(request),
  );

  ipcMain.handle(
    "reviews:comment:delete",
    async (
      _event,
      request: {
        reviewId: string;
        commentId: string;
      },
    ) => (await client()).deleteReviewComment(request),
  );
}
