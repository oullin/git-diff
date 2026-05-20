import { contextBridge, ipcRenderer } from "electron";

interface RunRequest {
  workflowId: string;
  confirmationOptionId: string;
  enabledPhaseIds: string[];
}

interface RunEvent {
  runId: string;
  seq: number;
  type: string;
  phaseId?: string;
  phaseName?: string;
  status?: string;
  message?: string;
}

interface RuntimeSettings {
  repoRoot: string;
  appsConfigPath: string;
  secretsConfigPath: string;
  generatedAppsPath: string;
  archiveRoot: string;
  workflowDbPath: string;
  opVault: string;
  opItem: string;
}

interface TemplateFileContent {
  file: {
    path: string;
    relative: string;
    kind: string;
    size: number;
    modifiedAt?: string;
    exists: boolean;
  };
  content: string;
}

interface OpVault {
  id: string;
  name: string;
}

interface OpItem {
  id: string;
  title: string;
}

interface AppDiagnostic {
  id: string;
  level: "info" | "warning" | "error";
  source: string;
  message: string;
  details?: string;
  createdAt: string;
}

type OpVaultsResult =
  | { ok: true; vaults: OpVault[] }
  | { ok: false; code: string; message: string };
type OpItemsResult = { ok: true; items: OpItem[] } | { ok: false; code: string; message: string };
type OpSigninResult = { ok: true } | { ok: false; message: string };
type OpInstallResult = { ok: true } | { ok: false; message: string };

contextBridge.exposeInMainWorld("macOS", {
  macName: () => ipcRenderer.invoke("system:macName"),
  macHostname: () => ipcRenderer.invoke("system:macHostname"),
  macSystemInfo: () => ipcRenderer.invoke("system:macSystemInfo"),
  workflows: () => ipcRenderer.invoke("workflows:list"),
  runs: (limit?: number) => ipcRenderer.invoke("runs:list", limit ?? 50),
  runLog: (runId: string) => ipcRenderer.invoke("runs:log", runId),
  templateFiles: () => ipcRenderer.invoke("template-files:list"),
  readTemplateFile: (path: string): Promise<TemplateFileContent> =>
    ipcRenderer.invoke("template-files:read", path),
  saveTemplateFile: (path: string, content: string): Promise<TemplateFileContent> =>
    ipcRenderer.invoke("template-files:save", path, content),
  settings: () => ipcRenderer.invoke("settings:get"),
  validateSettings: (settings: RuntimeSettings) =>
    ipcRenderer.invoke("settings:validate", settings),
  saveSettings: (settings: RuntimeSettings) => ipcRenderer.invoke("settings:save", settings),
  getUIPreferences: () => ipcRenderer.invoke("ui-prefs:get"),
  saveUIPreferences: (patch: Record<string, string>) => ipcRenderer.invoke("ui-prefs:save", patch),
  chooseDirectory: (defaultPath?: string) =>
    ipcRenderer.invoke("settings:choose-directory", defaultPath),
  chooseFile: (defaultPath?: string) => ipcRenderer.invoke("settings:choose-file", defaultPath),
  chooseSaveFile: (defaultPath?: string) =>
    ipcRenderer.invoke("settings:choose-save-file", defaultPath),
  listOpVaults: (): Promise<OpVaultsResult> => ipcRenderer.invoke("op:list-vaults"),
  listOpItems: (vault: string): Promise<OpItemsResult> =>
    ipcRenderer.invoke("op:list-items", vault),
  signinOpCli: (): Promise<OpSigninResult> => ipcRenderer.invoke("op:signin"),
  installOpDependencies: (): Promise<OpInstallResult> =>
    ipcRenderer.invoke("op:install-dependencies"),
  openDevTools: () => ipcRenderer.invoke("system:openDevTools"),
  appDiagnostics: (): Promise<AppDiagnostic[]> => ipcRenderer.invoke("diagnostics:list"),
  reportRendererError: (message: string, details?: string) =>
    ipcRenderer.invoke("diagnostics:renderer-error", { message, details }),
  onAppDiagnostic: (onEvent: (event: AppDiagnostic) => void) => {
    const listener = (_: Electron.IpcRendererEvent, event: AppDiagnostic) => onEvent(event);

    ipcRenderer.on("diagnostics:event", listener);

    return () => ipcRenderer.removeListener("diagnostics:event", listener);
  },
  runWorkflow: (request: RunRequest, onEvent: (event: RunEvent) => void) => {
    const channel = `workflow:event:${crypto.randomUUID()}`;
    const listener = (_: Electron.IpcRendererEvent, event: RunEvent) => onEvent(event);

    ipcRenderer.on(channel, listener);

    return ipcRenderer
      .invoke("workflow:run", request, channel)
      .finally(() => ipcRenderer.removeListener(channel, listener));
  },
});

contextBridge.exposeInMainWorld("diffApp", {
  takeLaunchIntent: () => ipcRenderer.invoke("launch-intent:take"),
  onLaunchIntent: (handler: (intent: unknown) => void) => {
    const listener = (_event: Electron.IpcRendererEvent, intent: unknown) => handler(intent);
    ipcRenderer.on("launch-intent:updated", listener);
    return () => ipcRenderer.removeListener("launch-intent:updated", listener);
  },
  repositoryState: (path?: string) => ipcRenderer.invoke("repository:state", path),
  openRepository: (path: string) => ipcRenderer.invoke("repository:open", path),
  refreshRepository: (path: string) => ipcRenderer.invoke("repository:refresh", path),
  readCommit: (sha: string, path?: string) => ipcRenderer.invoke("repository:commit", sha, path),
  listCommits: (path?: string, limit?: number) => ipcRenderer.invoke("repository:log", path, limit),
  readRepositoryFile: (root: string, path: string) =>
    ipcRenderer.invoke("repository:file:read", root, path),
  listBranches: (path?: string) => ipcRenderer.invoke("repository:branches", path),
  checkoutBranch: (path: string, branch: string) =>
    ipcRenderer.invoke("repository:checkout", path, branch),
  createBranch: (path: string, name: string) =>
    ipcRenderer.invoke("repository:branches:create", path, name),
  deleteBranch: (name: string, path?: string) =>
    ipcRenderer.invoke("repository:branches:delete", name, path),
  lockBranch: (name: string, path?: string) =>
    ipcRenderer.invoke("repository:branches:lock", name, path),
  unlockBranch: (name: string, path?: string) =>
    ipcRenderer.invoke("repository:branches:unlock", name, path),
  chooseRepository: (defaultPath?: string): Promise<string | null> =>
    ipcRenderer.invoke("repository:choose", defaultPath),
  listRepositories: () => ipcRenderer.invoke("repositories:list"),
  searchRepositoryFiles: (query: string, limit?: number) =>
    ipcRenderer.invoke("repositories:search-files", query, limit),
  upsertRepository: (request: { path: string; name?: string }) =>
    ipcRenderer.invoke("repositories:upsert", request),
  removeRepository: (path: string) => ipcRenderer.invoke("repositories:remove", path),
  listCollaborators: (path: string) => ipcRenderer.invoke("repositories:collaborators:list", path),
  addCollaborator: (request: { path: string; userId: number; role: "write" | "read" }) =>
    ipcRenderer.invoke("repositories:collaborators:add", request),
  removeCollaborator: (request: { path: string; userId: number }) =>
    ipcRenderer.invoke("repositories:collaborators:remove", request),
  createReview: (request: Record<string, unknown>) => ipcRenderer.invoke("reviews:create", request),
  listReviews: (limit?: number) => ipcRenderer.invoke("reviews:list", limit),
  reviewDetail: (id: string) => ipcRenderer.invoke("reviews:detail", id),
  addReviewEvent: (request: Record<string, unknown>) =>
    ipcRenderer.invoke("reviews:event", request),
  createReviewComment: (request: Record<string, unknown>) =>
    ipcRenderer.invoke("reviews:comment:create", request),
  updateReviewComment: (request: Record<string, unknown>) =>
    ipcRenderer.invoke("reviews:comment:update", request),
  deleteReviewComment: (request: Record<string, unknown>) =>
    ipcRenderer.invoke("reviews:comment:delete", request),
  getUIPreferences: () => ipcRenderer.invoke("ui-prefs:get"),
  saveUIPreferences: (patch: Record<string, string>) => ipcRenderer.invoke("ui-prefs:save", patch),
  getAuthState: () => ipcRenderer.invoke("auth:state"),
  authBootstrap: () => ipcRenderer.invoke("auth:bootstrap"),
  authSetup: (password: string) => ipcRenderer.invoke("auth:setup", { password }),
  authLogin: (password: string, remember: boolean) =>
    ipcRenderer.invoke("auth:login", { password, remember }),
  authLogout: () => ipcRenderer.invoke("auth:logout"),
  authWipe: (osUsername?: string) => ipcRenderer.invoke("auth:wipe", { osUsername }),
  openDevTools: () => ipcRenderer.invoke("system:openDevTools"),
  getSystemStats: () => ipcRenderer.invoke("system:stats"),
});
