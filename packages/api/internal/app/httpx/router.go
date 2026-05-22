package httpx

import "net/http"

func (s Server) BuildMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthz", s.healthz)
	mux.HandleFunc("GET /v1/system/stats", s.getSystemStats)

	mux.HandleFunc("GET /v1/repository/state", s.repositoryState)
	mux.HandleFunc("POST /v1/repository/open", s.repositoryOpen)
	mux.HandleFunc("POST /v1/repository/refresh", s.repositoryRefresh)
	mux.HandleFunc("GET /v1/repository/commit", s.repositoryCommit)
	mux.HandleFunc("GET /v1/repository/log", s.repositoryLog)
	mux.HandleFunc("GET /v1/repository/file", s.repositoryFile)
	mux.HandleFunc("GET /v1/repository/file-range", s.repositoryFileRange)
	mux.HandleFunc("GET /v1/repository/pull-requests", s.repositoryPullRequests)
	mux.HandleFunc("GET /v1/repository/pull-request", s.repositoryPullRequest)
	mux.HandleFunc("POST /v1/walkthrough", s.walkthroughGenerate)

	mux.HandleFunc("GET /v1/repository/branches", s.repositoryBranches)
	mux.HandleFunc("POST /v1/repository/checkout", s.repositoryCheckout)
	mux.HandleFunc("POST /v1/repository/branches/create", s.repositoryCreateBranch)
	mux.HandleFunc("POST /v1/repository/branches/lock", s.lockBranch)
	mux.HandleFunc("POST /v1/repository/branches/unlock", s.unlockBranch)
	mux.HandleFunc("DELETE /v1/repository/branches", s.deleteBranch)

	mux.HandleFunc("GET /v1/repositories", s.listRepositories)
	mux.HandleFunc("GET /v1/repositories/search-files", s.searchRepositoryFiles)
	mux.HandleFunc("POST /v1/repositories", s.upsertRepository)
	mux.HandleFunc("DELETE /v1/repositories", s.removeRepository)
	mux.HandleFunc("GET /v1/repositories/collaborators", s.listCollaborators)
	mux.HandleFunc("POST /v1/repositories/collaborators", s.addCollaborator)
	mux.HandleFunc("DELETE /v1/repositories/collaborators", s.removeCollaborator)

	mux.HandleFunc("POST /v1/reviews", s.createReview)
	mux.HandleFunc("GET /v1/reviews", s.listReviews)
	mux.HandleFunc("GET /v1/reviews/{id}", s.reviewDetail)
	mux.HandleFunc("POST /v1/reviews/{id}/events", s.addReviewEvent)
	mux.HandleFunc("POST /v1/reviews/{id}/comments", s.createReviewComment)
	mux.HandleFunc("PATCH /v1/reviews/{id}/comments/{commentId}", s.updateReviewComment)
	mux.HandleFunc("DELETE /v1/reviews/{id}/comments/{commentId}", s.deleteReviewComment)

	mux.HandleFunc("GET /v1/pending-comments", s.listPendingComments)
	mux.HandleFunc("POST /v1/pending-comments", s.createPendingComment)
	mux.HandleFunc("PATCH /v1/pending-comments/{id}", s.updatePendingComment)
	mux.HandleFunc("DELETE /v1/pending-comments/{id}", s.deletePendingComment)
	mux.HandleFunc("POST /v1/pending-comments/promote", s.promotePendingComments)

	mux.HandleFunc("GET /v1/preferences", s.getPreferences)
	mux.HandleFunc("POST /v1/preferences", s.savePreferences)

	mux.HandleFunc("GET /v1/userconfig", s.userConfigGet)
	mux.HandleFunc("GET /v1/userconfig/stream", s.userConfigStream)

	mux.HandleFunc("GET /v1/auth/state", s.authState)
	mux.HandleFunc("POST /v1/auth/setup", s.authSetup)
	mux.HandleFunc("POST /v1/auth/login", s.authLogin)
	mux.HandleFunc("POST /v1/auth/resume", s.authResume)
	mux.HandleFunc("POST /v1/auth/logout", s.authLogout)
	mux.HandleFunc("POST /v1/auth/wipe", s.authWipe)

	return mux
}
