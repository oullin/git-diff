package httpx

// Single source of truth for the HTTP endpoint paths served by BuildMux.
//
// Each path appears here once; router.go combines a const with its HTTP
// method when registering a handler. Wildcards ({id}, {commentId}) are part
// of the registered pattern and are read back via r.PathValue in handlers.
//
// This mirrors the TypeScript bridge routes (packages/bridge/src/routes.ts).
// The two are independent (no cross-language sharing); keep them in sync by
// hand when adding or renaming an endpoint.
const (
	routeHealthz     = "/v1/healthz"
	routeSystemStats = "/v1/system/stats"

	routeRepositoryState     = "/v1/repository/state"
	routeRepositoryOpen      = "/v1/repository/open"
	routeRepositoryRefresh   = "/v1/repository/refresh"
	routeRepositoryCommit    = "/v1/repository/commit"
	routeRepositoryLog       = "/v1/repository/log"
	routeRepositoryFile      = "/v1/repository/file"
	routeRepositoryFileRange = "/v1/repository/file-range"
	routeRepositoryFileRaw   = "/v1/repository/file/raw"
	routeRepositoryPullReqs  = "/v1/repository/pull-requests"
	routeRepositoryPullReq   = "/v1/repository/pull-request"
	routeWalkthrough         = "/v1/walkthrough"

	routeRepositoryBranches       = "/v1/repository/branches"
	routeRepositoryCheckout       = "/v1/repository/checkout"
	routeRepositoryBranchesCreate = "/v1/repository/branches/create"
	routeRepositoryBranchesLock   = "/v1/repository/branches/lock"
	routeRepositoryBranchesUnlock = "/v1/repository/branches/unlock"

	routeRepositories             = "/v1/repositories"
	routeRepositoriesSearchFiles  = "/v1/repositories/search-files"
	routeRepositoriesCollaborator = "/v1/repositories/collaborators"

	routeReviews              = "/v1/reviews"
	routeReviewDetail         = "/v1/reviews/{id}"
	routeReviewEvents         = "/v1/reviews/{id}/events"
	routeReviewComments       = "/v1/reviews/{id}/comments"
	routeReviewComment        = "/v1/reviews/{id}/comments/{commentId}"
	routeReviewCommentResolve = "/v1/reviews/{id}/comments/{commentId}/resolve"

	routePendingComments        = "/v1/pending-comments"
	routePendingComment         = "/v1/pending-comments/{id}"
	routePendingCommentsPromote = "/v1/pending-comments/promote"

	routePreferences = "/v1/preferences"

	routeUserConfig       = "/v1/userconfig"
	routeUserConfigStream = "/v1/userconfig/stream"

	routeAuthState  = "/v1/auth/state"
	routeAuthSetup  = "/v1/auth/setup"
	routeAuthLogin  = "/v1/auth/login"
	routeAuthResume = "/v1/auth/resume"
	routeAuthLogout = "/v1/auth/logout"
	routeAuthWipe   = "/v1/auth/wipe"
)
