import {
  type AuthStateResponse,
  type AuthUser,
  type RuntimeSettings,
  type SettingsResponse,
  type RunWorkflowRequest,
  type WorkflowEvent,
} from "@git-diff/bridge";
import {
  app,
  BrowserWindow,
  dialog,
  ipcMain,
  safeStorage,
  type OpenDialogOptions,
  type SaveDialogOptions,
} from "electron";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
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

  ipcMain.handle("repository:commit", async (_event, sha: string, path?: string) =>
    (await client()).readCommit({ path, sha }),
  );

  ipcMain.handle("repository:log", async (_event, path?: string, limit?: number) =>
    (await client()).listCommits({ path, limit }),
  );

  ipcMain.handle("repository:file:read", async (_event, root: string, path: string) =>
    (await client()).readRepositoryFile({ root, path }),
  );

  ipcMain.handle("repository:branches", async (_event, path?: string) =>
    (await client()).listBranches({ path }),
  );

  ipcMain.handle("repository:checkout", async (_event, path: string, branch: string) =>
    (await client()).checkoutBranch({ path, branch }),
  );

  ipcMain.handle("repository:branches:create", async (_event, path: string, name: string) =>
    (await client()).createBranch({ path, name }),
  );

  ipcMain.handle("repository:branches:delete", async (_event, name: string, path?: string) =>
    (await client()).deleteBranch({ path, name }),
  );

  ipcMain.handle("repository:branches:lock", async (_event, name: string, path?: string) =>
    (await client()).lockBranch({ path, name }),
  );

  ipcMain.handle("repository:branches:unlock", async (_event, name: string, path?: string) =>
    (await client()).unlockBranch({ path, name }),
  );

  ipcMain.handle("system:stats", async () => (await client()).getSystemStats());

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

  ipcMain.handle("repositories:search-files", async (_event, query: string, limit?: number) => {
    const response = await (await client()).searchRepositoryFiles({ query, limit });

    return response.results ?? [];
  });

  ipcMain.handle("repositories:upsert", async (_event, request: { path: string; name?: string }) =>
    (await client()).upsertRepository(request),
  );

  ipcMain.handle("repositories:remove", async (_event, path: string) =>
    (await client()).removeRepository({ path }),
  );

  ipcMain.handle("repositories:collaborators:list", async (_event, path: string) => {
    const response = await (await client()).listCollaborators({ path });
    return response.collaborators ?? [];
  });

  ipcMain.handle(
    "repositories:collaborators:add",
    async (_event, request: { path: string; userId: number; role: "write" | "read" }) =>
      (await client()).addCollaborator(request),
  );

  ipcMain.handle(
    "repositories:collaborators:remove",
    async (_event, request: { path: string; userId: number }) =>
      (await client()).removeCollaborator(request),
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

  ipcMain.handle("ui-prefs:get", async () => (await client()).getUIPreferences());

  ipcMain.handle("ui-prefs:save", async (_event, patch: Record<string, string>) =>
    (await client()).saveUIPreferences(patch ?? {}),
  );

  ipcMain.handle("auth:state", async () => (await client()).getAuthState());

  ipcMain.handle("auth:setup", async (_event, request: { password: string }) => {
    const response = await (await client()).authSetup(request);

    if (response.token) {
      writeSessionToken(response.token);
    }

    return response;
  });

  ipcMain.handle("auth:login", async (_event, request: { password: string; remember: boolean }) => {
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

  ipcMain.handle("auth:wipe", async (_event, request: { osUsername?: string } = {}) => {
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

function sessionTokenPath(): string {
  return path.join(app.getPath("userData"), ".session-token");
}

function readSessionToken(): string | null {
  const filePath = sessionTokenPath();

  if (!fs.existsSync(filePath)) {
    return null;
  }

  try {
    const buf = fs.readFileSync(filePath);

    if (safeStorage.isEncryptionAvailable()) {
      return safeStorage.decryptString(buf);
    }

    return buf.toString("utf8");
  } catch {
    try {
      fs.unlinkSync(filePath);
    } catch {
      /* ignore */
    }

    return null;
  }
}

function writeSessionToken(token: string): void {
  const filePath = sessionTokenPath();

  try {
    fs.mkdirSync(path.dirname(filePath), { recursive: true });

    if (safeStorage.isEncryptionAvailable()) {
      fs.writeFileSync(filePath, safeStorage.encryptString(token), { mode: 0o600 });
    } else {
      console.warn("electron safeStorage unavailable; persisting session token in plaintext");
      fs.writeFileSync(filePath, token, { mode: 0o600, encoding: "utf8" });
    }
  } catch (error) {
    console.warn("failed to persist session token", error);
  }
}

function clearSessionToken(): void {
  const filePath = sessionTokenPath();

  try {
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath);
    }
  } catch {
    /* ignore */
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
