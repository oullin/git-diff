import { contextBridge, ipcRenderer } from "electron";

contextBridge.exposeInMainWorld("diffApp", {
    takeLaunchIntent: () => ipcRenderer.invoke("launch-intent:take"),
    onLaunchIntent: (handler: (intent: unknown) => void) => {
        const listener = (_event: Electron.IpcRendererEvent, intent: unknown) => handler(intent);

        ipcRenderer.on("launch-intent:updated", listener);

        return () => ipcRenderer.removeListener("launch-intent:updated", listener);
    },
    openNewWindow: (repoPath?: string) => ipcRenderer.invoke("window:new", repoPath),
    repositoryState: (path?: string) => ipcRenderer.invoke("repository:state", path),
    openRepository: (path: string) => ipcRenderer.invoke("repository:open", path),
    refreshRepository: (path: string) => ipcRenderer.invoke("repository:refresh", path),
    readCommit: (sha: string, path?: string) => ipcRenderer.invoke("repository:commit", sha, path),
    listCommits: (path?: string, limit?: number) =>
        ipcRenderer.invoke("repository:log", path, limit),
    generateWalkthrough: (request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
        refresh?: boolean;
    }) => ipcRenderer.invoke("walkthrough:generate", request),
    listPullRequests: (path?: string, limit?: number) =>
        ipcRenderer.invoke("pull-requests:list", path, limit),
    readPullRequest: (number: number, path?: string) =>
        ipcRenderer.invoke("pull-requests:open", number, path),
    listPendingComments: (request: { path?: string; kind?: "working" | "commit"; sha?: string }) =>
        ipcRenderer.invoke("pending-comments:list", request),
    createPendingComment: (request: {
        repoRoot: string;
        contextKind: "working" | "commit";
        contextSha?: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        startLineNumber?: number;
        startSide?: string;
        authorLabel: string;
        bodyHtml: string;
    }) => ipcRenderer.invoke("pending-comments:create", request),
    updatePendingComment: (request: { id: number; bodyHtml: string }) =>
        ipcRenderer.invoke("pending-comments:update", request),
    deletePendingComment: (id: number) => ipcRenderer.invoke("pending-comments:delete", id),
    promotePendingComments: (reviewId: number) =>
        ipcRenderer.invoke("pending-comments:promote", reviewId),
    readRepositoryFile: (root: string, path: string) =>
        ipcRenderer.invoke("repository:file:read", root, path),
    readRepositoryFileRange: (request: {
        root: string;
        path: string;
        ref?: string;
        startLine: number;
        endLine: number;
    }) => ipcRenderer.invoke("repository:file:range", request),
    readRepositoryFileBytes: (request: { root?: string; path: string; ref?: string }) =>
        ipcRenderer.invoke("repository:file:bytes", request),
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
    listCollaborators: (path: string) =>
        ipcRenderer.invoke("repositories:collaborators:list", path),
    addCollaborator: (request: { path: string; userId: number; role: "write" | "read" }) =>
        ipcRenderer.invoke("repositories:collaborators:add", request),
    removeCollaborator: (request: { path: string; userId: number }) =>
        ipcRenderer.invoke("repositories:collaborators:remove", request),
    createReview: (request: Record<string, unknown>) =>
        ipcRenderer.invoke("reviews:create", request),
    listReviews: (limit?: number) => ipcRenderer.invoke("reviews:list", limit),
    reviewDetail: (id: number) => ipcRenderer.invoke("reviews:detail", id),
    addReviewEvent: (request: Record<string, unknown>) =>
        ipcRenderer.invoke("reviews:event", request),
    createReviewComment: (request: Record<string, unknown>) =>
        ipcRenderer.invoke("reviews:comment:create", request),
    updateReviewComment: (request: Record<string, unknown>) =>
        ipcRenderer.invoke("reviews:comment:update", request),
    deleteReviewComment: (request: Record<string, unknown>) =>
        ipcRenderer.invoke("reviews:comment:delete", request),
    getUIPreferences: () => ipcRenderer.invoke("ui-prefs:get"),
    saveUIPreferences: (patch: Record<string, string>) =>
        ipcRenderer.invoke("ui-prefs:save", patch),
    getUserConfig: () => ipcRenderer.invoke("user-config:get"),
    openUserConfigFile: () => ipcRenderer.invoke("user-config:open"),
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
