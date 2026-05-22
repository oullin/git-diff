import { client } from "#electron/bridge.js";
import type { IpcRouter } from "#electron/ipc/router.js";

export function register(router: IpcRouter): void {
    router.on("user-config:get", async () => (await client()).userConfig.get());
}
