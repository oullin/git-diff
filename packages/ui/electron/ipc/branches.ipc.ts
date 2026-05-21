import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle("repository:branches", async (_event, path?: string) =>
        (await client()).branches.list({ path }),
    );

    ipcMain.handle("repository:checkout", async (_event, path: string, branch: string) => {
        try {
            return await (await client()).branches.checkout({ path, branch });
        } catch (cause) {
            const err = cause as { code?: string; files?: string[]; message?: string };
            if (err && typeof err.code === "string") {
                const payload = JSON.stringify({
                    code: err.code,
                    files: Array.isArray(err.files) ? err.files : [],
                    message: typeof err.message === "string" ? err.message : "",
                });
                throw new Error(`__BRIDGE_ERROR__${payload}`);
            }
            throw cause;
        }
    });

    ipcMain.handle("repository:branches:create", async (_event, path: string, name: string) =>
        (await client()).branches.create({ path, name }),
    );

    ipcMain.handle("repository:branches:delete", async (_event, name: string, path?: string) =>
        (await client()).branches.delete({ path, name }),
    );

    ipcMain.handle("repository:branches:lock", async (_event, name: string, path?: string) =>
        (await client()).branches.lock({ path, name }),
    );

    ipcMain.handle("repository:branches:unlock", async (_event, name: string, path?: string) =>
        (await client()).branches.unlock({ path, name }),
    );
}
