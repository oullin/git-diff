import { client } from "#electron/bridge.js";
import type { IpcRouter } from "#electron/ipc/router.js";

export function register(router: IpcRouter): void {
    router.on("ui-prefs:get", async () => (await client()).preferences.get());

    router.on("ui-prefs:save", async (_event, patch: Record<string, string>) =>
        (await client()).preferences.save(patch ?? {}),
    );
}
