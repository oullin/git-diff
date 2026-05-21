import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";

export function register(): void {
    ipcMain.handle(
        "walkthrough:generate",
        async (
            _event,
            request: {
                path?: string;
                kind?: "working" | "commit";
                sha?: string;
                refresh?: boolean;
            },
        ) => (await client()).walkthroughs.generate(request),
    );
}
