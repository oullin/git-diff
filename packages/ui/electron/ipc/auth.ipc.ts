import { ipcMain } from "electron";
import type {
    AuthLoginRequest,
    AuthSetupRequest,
    AuthStateResponse,
    AuthUser,
    AuthWipeRequest,
} from "@git-diff/contracts";
import { client } from "#electron/bridge.js";
import {
    clearSessionToken,
    readSessionToken,
    writeSessionToken,
} from "#electron/ipc/session-token.js";

export function register(): void {
    ipcMain.handle("auth:state", async () => (await client()).getAuthState());

    ipcMain.handle("auth:setup", async (_event, request: AuthSetupRequest) => {
        const response = await (await client()).authSetup(request);

        if (response.token) {
            writeSessionToken(response.token);
        }

        return response;
    });

    ipcMain.handle("auth:login", async (_event, request: AuthLoginRequest) => {
        const response = await (await client()).authLogin(request);

        if (request.remember && response.token) {
            writeSessionToken(response.token);
        } else {
            clearSessionToken();
        }

        return response;
    });

    ipcMain.handle("auth:logout", async () => {
        clearSessionToken();

        await (await client()).authLogout();
    });

    ipcMain.handle("auth:wipe", async (_event, request: AuthWipeRequest = {}) => {
        clearSessionToken();

        await (await client()).authWipe({ osUsername: request.osUsername });
    });

    ipcMain.handle(
        "auth:bootstrap",
        async (): Promise<{ user: AuthUser | null; state: AuthStateResponse }> => {
            const token = readSessionToken();
            const c = await client();
            let user: AuthUser | null = null;

            if (token) {
                try {
                    const resumed = await c.authResume({ token });
                    user = resumed.user;
                } catch {
                    clearSessionToken();
                }
            }

            const state = await c.getAuthState();

            return { user, state };
        },
    );
}
