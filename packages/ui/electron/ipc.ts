import {
  type RuntimeSettings,
  type SettingsResponse,
  type RunWorkflowRequest,
  type WorkflowEvent,
} from "@git-diff/bridge";
import {
  BrowserWindow,
  dialog,
  ipcMain,
  type OpenDialogOptions,
  type SaveDialogOptions,
} from "electron";
import os from "node:os";
import {
  client,
  getBridgeSettings,
  hasExternalBridge,
  setBridgeSettings,
  startBridgeIfNeeded,
  stopWorkflowBridge,
} from "#electron/bridge.js";
import { moveWorkflowDatabase, writeSavedSettings } from "#electron/settings-store.js";
import { accountAvatarUrl, architectureLabel, osLabel } from "#electron/system-info.js";
import { openTerminalCommand } from "#electron/terminal.js";

type IpcDeps = {
  getMainWindow: () => BrowserWindow | null;
  openDevToolsPanel: (window: BrowserWindow) => void;
};

export function registerIpcHandlers(deps: IpcDeps) {
  ipcMain.handle("repository:state", async (_event, path?: string) =>
    (await client()).repositoryState({ path }),
  );

  ipcMain.handle("repository:open", async (_event, path: string) =>
    (await client()).openRepository({ path }),
  );

  ipcMain.handle("repository:refresh", async (_event, path: string) =>
    (await client()).refreshRepository({ path }),
  );

  ipcMain.handle("reviews:create", async (_event, request: Record<string, unknown>) =>
    (await client()).createReview(request),
  );

  ipcMain.handle("reviews:list", async (_event, limit?: number) =>
    (await client()).listReviews({ limit }),
  );

  ipcMain.handle("reviews:detail", async (_event, id: string) =>
    (await client()).reviewDetail({ id }),
  );

  ipcMain.handle(
    "reviews:event",
    async (
      _event,
      request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
      },
    ) => (await client()).addReviewEvent(request),
  );

  ipcMain.handle(
    "reviews:comment:create",
    async (
      _event,
      request: {
        reviewId: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        authorLabel: string;
        bodyHtml: string;
      },
    ) => (await client()).createReviewComment(request),
  );

  ipcMain.handle(
    "reviews:comment:update",
    async (
      _event,
      request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
      },
    ) => (await client()).updateReviewComment(request),
  );

  ipcMain.handle(
    "reviews:comment:delete",
    async (
      _event,
      request: {
        reviewId: string;
        commentId: string;
      },
    ) => (await client()).deleteReviewComment(request),
  );

  ipcMain.handle("repositories:list", async () => {
    const response = await (await client()).listRepositories();

    return response.repositories ?? [];
  });

  ipcMain.handle("repositories:upsert", async (_event, request: { path: string; name?: string }) =>
    (await client()).upsertRepository(request),
  );

  ipcMain.handle("repositories:remove", async (_event, path: string) =>
    (await client()).removeRepository({ path }),
  );

  ipcMain.handle("repository:choose", async (_event, defaultPath?: string) => {
    const options: OpenDialogOptions = {
      defaultPath,
      properties: ["openDirectory"],
    };
    const mainWindow = deps.getMainWindow();
    const result = mainWindow
      ? await dialog.showOpenDialog(mainWindow, options)
      : await dialog.showOpenDialog(options);

    return result.canceled ? null : (result.filePaths[0] ?? null);
  });

  ipcMain.handle("workflows:list", async () => {
    const response = await (await client()).listWorkflows();

    return response.workflows ?? [];
  });

  ipcMain.handle("runs:list", async (_event, limit: number) => {
    const response = await (await client()).listRuns({ limit });

    return response.runs ?? [];
  });

  ipcMain.handle("runs:log", async (_event, runId: string) => (await client()).runLog({ runId }));

  ipcMain.handle("template-files:list", async () => {
    const response = await (await client()).listTemplateFiles();

    return response.files ?? [];
  });

  ipcMain.handle("template-files:read", async (_event, path: string) =>
    (await client()).readTemplateFile({ path }),
  );

  ipcMain.handle("template-files:save", async (_event, path: string, content: string) =>
    (await client()).saveTemplateFile({ path, content }),
  );

  ipcMain.handle("settings:get", async () => (await client()).getSettings());

  ipcMain.handle("settings:validate", async (_event, settings: RuntimeSettings) =>
    (await client()).validateSettings({ settings }),
  );

  ipcMain.handle("settings:save", async (_event, settings: RuntimeSettings) =>
    saveSettings(settings),
  );

  ipcMain.handle("preferences:get", async () => (await client()).getUserPreferences());

  ipcMain.handle(
    "preferences:save",
    async (_event, preferences: Record<string, unknown> | string) =>
      typeof preferences === "string"
        ? (await client()).saveUserPreferences({ theme: preferences })
        : (await client()).saveUserPreferences(preferences),
  );

  ipcMain.handle("op:list-vaults", async () => {
    try {
      const response = await (await client()).listOpVaults();

      return { ok: true as const, vaults: response.vaults ?? [] };
    } catch (error) {
      return opErrorEnvelope(error);
    }
  });

  ipcMain.handle("op:list-items", async (_event, vault: string) => {
    try {
      const response = await (await client()).listOpItems({ vault });

      return { ok: true as const, items: response.items ?? [] };
    } catch (error) {
      return opErrorEnvelope(error);
    }
  });

  ipcMain.handle("op:signin", async () => {
    return openTerminalCommand(
      'op signin && echo "\\n[Signed in. You can close this window and return to Gus’ MacBook Setup.]"',
    );
  });

  ipcMain.handle("op:install-dependencies", async () => {
    return openTerminalCommand(
      [
        'if ! command -v brew >/dev/null 2>&1; then echo "Homebrew is required. Run ./setup.sh first, then retry."; exit 1; fi',
        "brew install --cask 1password 1password-cli",
        'echo "\\n[1Password and 1Password CLI install finished. Open 1Password, enable CLI integration if needed, then return to Gus’ MacBook Setup.]"',
      ].join("; "),
    );
  });

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

  ipcMain.handle("system:macName", () => os.userInfo().username);
  ipcMain.handle("system:macHostname", () => os.hostname());
  ipcMain.handle("system:macSystemInfo", () => ({
    name: os.userInfo().username,
    hostname: os.hostname(),
    osLabel: osLabel(),
    architectureLabel: architectureLabel(os.arch()),
    avatarUrl: accountAvatarUrl(),
  }));
  ipcMain.handle("system:openDevTools", () => {
    const mainWindow = deps.getMainWindow();

    if (!mainWindow) {
      return;
    }

    deps.openDevToolsPanel(mainWindow);
  });

  ipcMain.handle(
    "workflow:run",
    async (event, request: RunWorkflowRequest, eventChannel: string) => {
      const c = await client();

      return new Promise<{ exitCode: number }>((resolveResult, reject) => {
        const stream = c.runWorkflow(request);
        let exitCode = 0;

        stream.on("data", (workflowEvent: WorkflowEvent) => {
          if (workflowEvent.type === "run_failed") {
            exitCode = 1;
          }

          event.sender.send(eventChannel, workflowEvent);
        });

        stream.on("error", reject);
        stream.on("end", () => resolveResult({ exitCode }));
      });
    },
  );
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

function opErrorEnvelope(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  const code =
    error &&
    typeof error === "object" &&
    "code" in error &&
    typeof (error as { code: unknown }).code === "string"
      ? (error as { code: string }).code
      : "op_failed";

  return { ok: false as const, code, message };
}
