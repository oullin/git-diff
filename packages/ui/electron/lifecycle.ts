import { app, dialog } from "electron";
import { appIcon } from "#electron/app-icon.js";
import { setBridgeSettings, startBridgeIfNeeded, stopWorkflowBridge } from "#electron/bridge.js";
import { recordDiagnostic } from "#electron/diagnostics.js";
import { registerIpcHandlers } from "#electron/ipc.js";
import type { LaunchIntent } from "#electron/launch-intent.js";
import { parseLaunchArgs } from "#electron/launch-intent.js";
import { installApplicationMenu } from "#electron/menu.js";
import { readSavedSettings } from "#electron/settings-store.js";
import {
    createWindow,
    getMainWindow,
    listAppWindows,
    openDevToolsPanel,
    setIntentForWindow,
} from "#electron/windows.js";

/**
 * runApp wires every Electron-lifecycle event for a normal launch. The
 * caller resolves the initial intent (working tree, commit, PR) and hands
 * it in; this module owns the single-instance lock, the whenReady chain
 * (icon, settings, IPC, menu, window, bridge), and the quit/activate
 * handlers.
 */
export function runApp(initialIntent: LaunchIntent): void {
    const singleInstanceLock = app.requestSingleInstanceLock();

    if (!singleInstanceLock) {
        app.quit();

        return;
    }

    app.on("second-instance", (_event, argv, workingDir) => {
        const intent = parseLaunchArgs(argv, app.isPackaged, workingDir || process.cwd());

        if (intent.kind === "help") {
            dialog.showMessageBox({
                type: "info",
                message: "git-diff",
                detail: intent.helpText ?? "",
            });

            return;
        }

        // Every second-instance launch opens a fresh window so two `git-diff`
        // invocations from two terminals end up side-by-side.
        const window = createWindow(intent);
        setIntentForWindow(window, intent);
        window.webContents.once("did-finish-load", () => {
            window.webContents.send("launch-intent:updated", intent);
        });
    });

    app.whenReady().then(() => {
        try {
            applyDockIcon();
            setBridgeSettings(readSavedSettings());
            registerIpcHandlers({ getMainWindow, openDevToolsPanel });
            installApplicationMenu();
            createWindow(initialIntent);
        } catch (error) {
            recordDiagnostic({
                level: "error",
                source: "Main process",
                message: error instanceof Error ? error.message : String(error),
                details: error instanceof Error ? error.stack : undefined,
            });
            console.error(error);
            app.quit();

            return;
        }

        void startBridgeIfNeeded().catch((error: unknown) => {
            recordDiagnostic({
                level: "error",
                source: "Backend bridge",
                message: error instanceof Error ? error.message : String(error),
                details: error instanceof Error ? error.stack : undefined,
            });
            console.error("Failed to start api HTTP bridge", error);
        });
    });

    app.on("window-all-closed", () => {
        if (process.platform !== "darwin") {
            app.quit();
        }
    });

    app.on("activate", () => {
        // On macOS, dock click reopens a window only if none are around.
        if (listAppWindows().length === 0) {
            createWindow();
        }
    });

    app.on("before-quit", stopWorkflowBridge);
}

export function showHelpAndExit(helpText: string): void {
    process.stderr.write(helpText + "\n");
    app.whenReady()
        .then(() => {
            dialog.showMessageBoxSync({
                type: "info",
                message: "git-diff",
                detail: helpText,
            });
        })
        .finally(() => app.exit(0));
}

function applyDockIcon(): void {
    const icon = appIcon();

    if (icon && process.platform === "darwin" && app.dock) {
        app.dock.setIcon(icon);
    }
}
