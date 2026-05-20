import { dialog, ipcMain, type OpenDialogOptions, type SaveDialogOptions } from "electron";
import type { RuntimeSettings, SettingsResponse } from "@git-diff/contracts";
import {
  client,
  getBridgeSettings,
  hasExternalBridge,
  setBridgeSettings,
  startBridgeIfNeeded,
  stopWorkflowBridge,
} from "#electron/bridge.js";
import { moveWorkflowDatabase, writeSavedSettings } from "#electron/settings-store.js";
import type { IpcDeps } from "#electron/ipc/types.js";

export function register(deps: IpcDeps): void {
  ipcMain.handle("settings:get", async () => (await client()).getSettings());

  ipcMain.handle("settings:validate", async (_event, settings: RuntimeSettings) =>
    (await client()).validateSettings({ settings }),
  );

  ipcMain.handle("settings:save", async (_event, settings: RuntimeSettings) =>
    saveSettings(settings),
  );

  ipcMain.handle("ui-prefs:get", async () => (await client()).getUIPreferences());

  ipcMain.handle("ui-prefs:save", async (_event, patch: Record<string, string>) =>
    (await client()).saveUIPreferences(patch ?? {}),
  );

  ipcMain.handle("settings:choose-directory", async (_event, defaultPath?: string) => {
    const options: OpenDialogOptions = {
      defaultPath,
      properties: ["openDirectory", "createDirectory"],
    };
    const mainWindow = deps.getMainWindow();
    const result = mainWindow
      ? await dialog.showOpenDialog(mainWindow, options)
      : await dialog.showOpenDialog(options);

    return result.canceled ? null : (result.filePaths[0] ?? null);
  });

  ipcMain.handle("settings:choose-file", async (_event, defaultPath?: string) => {
    const options: OpenDialogOptions = {
      defaultPath,
      properties: ["openFile"],
    };
    const mainWindow = deps.getMainWindow();
    const result = mainWindow
      ? await dialog.showOpenDialog(mainWindow, options)
      : await dialog.showOpenDialog(options);

    return result.canceled ? null : (result.filePaths[0] ?? null);
  });

  ipcMain.handle("settings:choose-save-file", async (_event, defaultPath?: string) => {
    const options: SaveDialogOptions = {
      defaultPath,
      properties: ["createDirectory", "showOverwriteConfirmation"],
    };
    const mainWindow = deps.getMainWindow();
    const result = mainWindow
      ? await dialog.showSaveDialog(mainWindow, options)
      : await dialog.showSaveDialog(options);

    return result.canceled ? null : (result.filePath ?? null);
  });
}

async function saveSettings(settings: RuntimeSettings): Promise<SettingsResponse> {
  const validation = await (await client()).validateSettings({ settings });

  if (!validation.valid || !validation.settings) {
    return validation;
  }

  const current = await (await client()).getSettings();
  const previousSettings = getBridgeSettings();
  const nextSettings = validation.settings;
  let rollbackDatabaseMove = () => {};
  let bridgeStopped = false;

  try {
    if (!hasExternalBridge()) {
      stopWorkflowBridge();
      bridgeStopped = true;
    }

    rollbackDatabaseMove = moveWorkflowDatabase(
      current.settings?.workflowDbPath,
      nextSettings.workflowDbPath,
    );
    setBridgeSettings(nextSettings);
    writeSavedSettings(nextSettings);

    if (hasExternalBridge()) {
      return validation;
    }

    return await (await client()).getSettings();
  } catch (error) {
    rollbackDatabaseMove();
    setBridgeSettings(previousSettings);
    writeSavedSettings(previousSettings);

    if (bridgeStopped) {
      await startBridgeIfNeeded();
    }

    throw error;
  }
}
