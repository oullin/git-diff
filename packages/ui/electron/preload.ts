import { contextBridge, ipcRenderer } from 'electron';
import { IpcChannels } from '#electron/ipc/routes.js';

contextBridge.exposeInMainWorld('diffApp', {
	takeLaunchIntent: () => ipcRenderer.invoke(IpcChannels.launchIntent.take),
	onLaunchIntent: (handler: (intent: unknown) => void) => {
		const listener = (_event: Electron.IpcRendererEvent, intent: unknown) => handler(intent);

		ipcRenderer.on(IpcChannels.launchIntent.updated, listener);

		return () => ipcRenderer.removeListener(IpcChannels.launchIntent.updated, listener);
	},
	openNewWindow: (repoPath?: string) => ipcRenderer.invoke(IpcChannels.window.new, repoPath),
	repositoryState: (path?: string) => ipcRenderer.invoke(IpcChannels.repository.state, path),
	openRepository: (path: string) => ipcRenderer.invoke(IpcChannels.repository.open, path),
	refreshRepository: (path: string) => ipcRenderer.invoke(IpcChannels.repository.refresh, path),
	readCommit: (sha: string, path?: string) => ipcRenderer.invoke(IpcChannels.repository.commit, sha, path),
	listCommits: (path?: string, limit?: number) => ipcRenderer.invoke(IpcChannels.repository.log, path, limit),
	generateWalkthrough: (request: { path?: string; kind?: 'working' | 'commit'; sha?: string; refresh?: boolean }) => ipcRenderer.invoke(IpcChannels.walkthrough.generate, request),
	listPullRequests: (path?: string, limit?: number) => ipcRenderer.invoke(IpcChannels.pullRequests.list, path, limit),
	readPullRequest: (number: number, path?: string) => ipcRenderer.invoke(IpcChannels.pullRequests.open, number, path),
	listPendingComments: (request: { path?: string; kind?: 'working' | 'commit'; sha?: string }) => ipcRenderer.invoke(IpcChannels.pendingComments.list, request),
	createPendingComment: (request: {
		repoRoot: string;
		contextKind: 'working' | 'commit';
		contextSha?: string;
		filePath: string;
		diffSection: string;
		side: string;
		lineNumber: number;
		startLineNumber?: number;
		startSide?: string;
		authorLabel: string;
		bodyHtml: string;
	}) => ipcRenderer.invoke(IpcChannels.pendingComments.create, request),
	updatePendingComment: (request: { id: number; bodyHtml: string }) => ipcRenderer.invoke(IpcChannels.pendingComments.update, request),
	deletePendingComment: (id: number) => ipcRenderer.invoke(IpcChannels.pendingComments.delete, id),
	promotePendingComments: (reviewId: number) => ipcRenderer.invoke(IpcChannels.pendingComments.promote, reviewId),
	readRepositoryFile: (root: string, path: string) => ipcRenderer.invoke(IpcChannels.repository.fileRead, root, path),
	readRepositoryFileRange: (request: { root: string; path: string; ref?: string; startLine: number; endLine: number }) => ipcRenderer.invoke(IpcChannels.repository.fileRange, request),
	readRepositoryFileBytes: (request: { root?: string; path: string; ref?: string }) => ipcRenderer.invoke(IpcChannels.repository.fileBytes, request),
	listBranches: (path?: string) => ipcRenderer.invoke(IpcChannels.branches.list, path),
	checkoutBranch: (path: string, branch: string) => ipcRenderer.invoke(IpcChannels.branches.checkout, path, branch),
	createBranch: (path: string, name: string) => ipcRenderer.invoke(IpcChannels.branches.create, path, name),
	deleteBranch: (name: string, path?: string) => ipcRenderer.invoke(IpcChannels.branches.delete, name, path),
	lockBranch: (name: string, path?: string) => ipcRenderer.invoke(IpcChannels.branches.lock, name, path),
	unlockBranch: (name: string, path?: string) => ipcRenderer.invoke(IpcChannels.branches.unlock, name, path),
	chooseRepository: (defaultPath?: string): Promise<string | null> => ipcRenderer.invoke(IpcChannels.repository.choose, defaultPath),
	listRepositories: () => ipcRenderer.invoke(IpcChannels.repositories.list),
	searchRepositoryFiles: (query: string, limit?: number) => ipcRenderer.invoke(IpcChannels.repositories.searchFiles, query, limit),
	upsertRepository: (request: { path: string; name?: string }) => ipcRenderer.invoke(IpcChannels.repositories.upsert, request),
	removeRepository: (path: string) => ipcRenderer.invoke(IpcChannels.repositories.remove, path),
	listCollaborators: (path: string) => ipcRenderer.invoke(IpcChannels.repositories.collaboratorsList, path),
	addCollaborator: (request: { path: string; userId: number; role: 'write' | 'read' }) => ipcRenderer.invoke(IpcChannels.repositories.collaboratorsAdd, request),
	removeCollaborator: (request: { path: string; userId: number }) => ipcRenderer.invoke(IpcChannels.repositories.collaboratorsRemove, request),
	createReview: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.create, request),
	listReviews: (limit?: number) => ipcRenderer.invoke(IpcChannels.reviews.list, limit),
	reviewDetail: (id: number) => ipcRenderer.invoke(IpcChannels.reviews.detail, id),
	addReviewEvent: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.event, request),
	createReviewComment: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.commentCreate, request),
	updateReviewComment: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.commentUpdate, request),
	deleteReviewComment: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.commentDelete, request),
	setReviewCommentResolved: (request: Record<string, unknown>) => ipcRenderer.invoke(IpcChannels.reviews.commentResolve, request),
	getUIPreferences: () => ipcRenderer.invoke(IpcChannels.preferences.get),
	saveUIPreferences: (patch: Record<string, string>) => ipcRenderer.invoke(IpcChannels.preferences.save, patch),
	getUserConfig: () => ipcRenderer.invoke(IpcChannels.userConfig.get),
	openUserConfigFile: () => ipcRenderer.invoke(IpcChannels.userConfig.open),
	getAuthState: () => ipcRenderer.invoke(IpcChannels.auth.state),
	authBootstrap: () => ipcRenderer.invoke(IpcChannels.auth.bootstrap),
	authSetup: (password: string) => ipcRenderer.invoke(IpcChannels.auth.setup, { password }),
	authLogin: (password: string, remember: boolean) => ipcRenderer.invoke(IpcChannels.auth.login, { password, remember }),
	authLogout: () => ipcRenderer.invoke(IpcChannels.auth.logout),
	authWipe: (osUsername?: string) => ipcRenderer.invoke(IpcChannels.auth.wipe, { osUsername }),
	openDevTools: () => ipcRenderer.invoke(IpcChannels.system.openDevTools),
	getSystemStats: () => ipcRenderer.invoke(IpcChannels.system.stats),
});
