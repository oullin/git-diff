// Single source of truth for the backend HTTP endpoint paths.
//
// Each `/v1/*` path lives here exactly once so a route rename is a one-line
// change instead of a hunt across the client classes. Static paths are plain
// strings; parameterized paths are builder functions. The routes hold the
// path only — clients still append their own query strings.
//
// This mirrors the Go router (packages/api/internal/app/httpx/routes.go).
// The two are independent (no cross-language sharing); keep them in sync by
// hand when adding or renaming an endpoint.
export const HttpRoutes = {
	system: {
		healthz: '/v1/healthz',
		stats: '/v1/system/stats',
	},
	repository: {
		state: '/v1/repository/state',
		open: '/v1/repository/open',
		refresh: '/v1/repository/refresh',
		commit: '/v1/repository/commit',
		log: '/v1/repository/log',
		file: '/v1/repository/file',
		fileRaw: '/v1/repository/file/raw',
		fileRange: '/v1/repository/file-range',
	},
	pullRequests: {
		list: '/v1/repository/pull-requests',
		read: '/v1/repository/pull-request',
	},
	branches: {
		// `list` doubles as the DELETE target.
		list: '/v1/repository/branches',
		checkout: '/v1/repository/checkout',
		create: '/v1/repository/branches/create',
		lock: '/v1/repository/branches/lock',
		unlock: '/v1/repository/branches/unlock',
	},
	walkthrough: {
		generate: '/v1/walkthrough',
	},
	repositories: {
		// `base` serves GET (list), POST (upsert) and DELETE (remove).
		base: '/v1/repositories',
		searchFiles: '/v1/repositories/search-files',
		collaborators: '/v1/repositories/collaborators',
	},
	reviews: {
		base: '/v1/reviews',
		detail: (id: number) => `/v1/reviews/${id}`,
		events: (id: number) => `/v1/reviews/${id}/events`,
		comments: (id: number) => `/v1/reviews/${id}/comments`,
		comment: (id: number, commentId: number) => `/v1/reviews/${id}/comments/${commentId}`,
		commentResolve: (id: number, commentId: number) => `/v1/reviews/${id}/comments/${commentId}/resolve`,
	},
	pendingComments: {
		base: '/v1/pending-comments',
		item: (id: number) => `/v1/pending-comments/${id}`,
		promote: '/v1/pending-comments/promote',
	},
	preferences: {
		base: '/v1/preferences',
	},
	userConfig: {
		base: '/v1/userconfig',
		// Hot-reload SSE, consumed directly via EventSource.
		stream: '/v1/userconfig/stream',
	},
	auth: {
		state: '/v1/auth/state',
		setup: '/v1/auth/setup',
		login: '/v1/auth/login',
		resume: '/v1/auth/resume',
		logout: '/v1/auth/logout',
		wipe: '/v1/auth/wipe',
	},
} as const;
