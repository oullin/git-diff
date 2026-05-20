import { app, dialog } from "electron";
import { appIcon } from "#electron/app-icon.js";
import { setBridgeSettings, startBridgeIfNeeded, stopWorkflowBridge } from "#electron/bridge.js";
import { recordDiagnostic, registerDiagnosticsIpc } from "#electron/diagnostics.js";
import { registerIpcHandlers } from "#electron/ipc.js";
import { setLaunchIntent } from "#electron/launch-intent-store.js";
import { parseLaunchArgs } from "#electron/launch-intent.js";
import { readSavedSettings } from "#electron/settings-store.js";
import {
  createWindow,
  focusMainWindow,
  getMainWindow,
  openDevToolsPanel,
} from "#electron/windows.js";

const initialIntent = parseLaunchArgs(process.argv, app.isPackaged, process.cwd());

if (initialIntent.kind === "help") {
  // Print to stderr so terminal users see it immediately, then surface a small
  // dialog for double-click launches. Quit either way.
  process.stderr.write((initialIntent.helpText ?? "") + "\n");
  app
    .whenReady()
    .then(() => {
      dialog.showMessageBoxSync({
        type: "info",
        message: "git-diff",
        detail: initialIntent.helpText ?? "",
      });
    })
    .finally(() => app.exit(0));
} else {
  setLaunchIntent(initialIntent);

  const singleInstanceLock = app.requestSingleInstanceLock();

  if (!singleInstanceLock) {
    app.quit();
  } else {
    app.on("second-instance", (_event, argv, workingDir) => {
      const intent = parseLaunchArgs(argv, app.isPackaged, workingDir || process.cwd());
      if (intent.kind === "help") {
        dialog.showMessageBox({ type: "info", message: "git-diff", detail: intent.helpText ?? "" });
        return;
      }
      setLaunchIntent(intent);
      focusMainWindow();
      const window = getMainWindow();
      window?.webContents.send("launch-intent:updated", intent);
    });
  }

  app.whenReady().then(() => {
    try {
      const icon = appIcon();
      if (icon && process.platform === "darwin" && app.dock) {
        app.dock.setIcon(icon);
      }

      setBridgeSettings(readSavedSettings());
      registerDiagnosticsIpc();
      registerIpcHandlers({ getMainWindow, openDevToolsPanel });
      createWindow();
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
    createWindow();
  });

  app.on("before-quit", stopWorkflowBridge);
}
