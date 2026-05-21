import type {
    AuthLoginRequest,
    AuthSetupRequest,
    AuthStateResponse,
    AuthUser,
    AuthWipeRequest,
} from "@git-diff/contracts";
import { client } from "#electron/bridge.js";
import type { IpcRouter } from "#electron/ipc/router.js";
import {
    clearSessionToken,
    readSessionToken,
    writeSessionToken,
} from "#electron/ipc/session-token.js";

export function register(router: IpcRouter): void {
    router.on("auth:state", async () => (await client()).auth.getState());

    router.on("auth:setup", async (_event, request: AuthSetupRequest) => {
        const response = await (await client()).auth.setup(request);

        if (response.token) {
            writeSessionToken(response.token);
        }

        return response;
    });

    router.on("auth:login", async (_event, request: AuthLoginRequest) => {
        const response = await (await client()).auth.login(request);

        if (request.remember && response.token) {
            writeSessionToken(response.token);
        } else {
            clearSessionToken();
        }

        return response;
    });

    router.on("auth:logout", async () => {
        clearSessionToken();

        await (await client()).auth.logout();
    });

    router.on("auth:wipe", async (_event, request: AuthWipeRequest = {}) => {
        clearSessionToken();

        await (await client()).auth.wipe({ osUsername: request.osUsername });
    });

    router.on(
        "auth:bootstrap",
        async (): Promise<{ user: AuthUser | null; state: AuthStateResponse }> => {
            const token = readSessionToken();
            const c = await client();
            let user: AuthUser | null = null;

            if (token) {
                try {
                    const resumed = await c.auth.resume({ token });
                    user = resumed.user;
                } catch {
                    clearSessionToken();
                }
            }

            const state = await c.auth.getState();

            return { user, state };
        },
    );
}
