import { client } from "#electron/bridge.js";
import type { IpcRouter } from "#electron/ipc/router.js";

export function register(router: IpcRouter): void {
    router.on("pull-requests:list", async (_event, path?: string, limit?: number) =>
        (await client()).pullRequests.list({ path, limit }),
    );

    router.on("pull-requests:open", async (_event, number: number, path?: string) =>
        (await client()).pullRequests.read({ path, number }),
    );
}
