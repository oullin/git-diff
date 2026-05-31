// Single source of truth for Electron IPC channel names.
//
// Every channel string lives here exactly once. Handler registration
// (`router.on(...)` in the `*.ipc.ts` files) and the renderer bridge
// (`ipcRenderer.invoke(...)` / `ipcRenderer.on(...)` in preload.ts) both
// reference these constants, so a rename is a single-line change and a typo
// surfaces as a TypeScript error rather than a silently broken channel.
//
// Grouping is logical, not prefix-driven: branch channels keep their
// `repository:` wire prefix but live under `branches` for readability.
export const IpcChannels = {
	repository: {
		state: 'repository:state',
		open: 'repository:open',
		refresh: 'repository:refresh',
		commit: 'repository:commit',
		log: 'repository:log',
		fileRead: 'repository:file:read',
		fileRange: 'repository:file:range',
		fileBytes: 'repository:file:bytes',
		choose: 'repository:choose',
	},
	branches: {
		list: 'repository:branches',
		checkout: 'repository:checkout',
		create: 'repository:branches:create',
		delete: 'repository:branches:delete',
		lock: 'repository:branches:lock',
		unlock: 'repository:branches:unlock',
	},
	repositories: {
		list: 'repositories:list',
		searchFiles: 'repositories:search-files',
		upsert: 'repositories:upsert',
		remove: 'repositories:remove',
		collaboratorsList: 'repositories:collaborators:list',
		collaboratorsAdd: 'repositories:collaborators:add',
		collaboratorsRemove: 'repositories:collaborators:remove',
	},
	pullRequests: {
		list: 'pull-requests:list',
		open: 'pull-requests:open',
	},
	pendingComments: {
		list: 'pending-comments:list',
		create: 'pending-comments:create',
		update: 'pending-comments:update',
		delete: 'pending-comments:delete',
		promote: 'pending-comments:promote',
	},
	reviews: {
		create: 'reviews:create',
		list: 'reviews:list',
		detail: 'reviews:detail',
		event: 'reviews:event',
		commentCreate: 'reviews:comment:create',
		commentUpdate: 'reviews:comment:update',
		commentDelete: 'reviews:comment:delete',
		commentResolve: 'reviews:comment:resolve',
	},
	walkthrough: {
		generate: 'walkthrough:generate',
	},
	preferences: {
		get: 'ui-prefs:get',
		save: 'ui-prefs:save',
	},
	userConfig: {
		get: 'user-config:get',
		open: 'user-config:open',
	},
	auth: {
		state: 'auth:state',
		bootstrap: 'auth:bootstrap',
		setup: 'auth:setup',
		login: 'auth:login',
		logout: 'auth:logout',
		wipe: 'auth:wipe',
	},
	system: {
		openDevTools: 'system:openDevTools',
		stats: 'system:stats',
	},
	launchIntent: {
		take: 'launch-intent:take',
		// Main -> renderer push event (webContents.send / ipcRenderer.on).
		updated: 'launch-intent:updated',
	},
	window: {
		new: 'window:new',
	},
} as const;
