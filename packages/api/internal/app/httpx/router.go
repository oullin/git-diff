package httpx

import "net/http"

func (s Server) BuildMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+routeHealthz, s.healthz)
	mux.HandleFunc("GET "+routeSystemStats, s.getSystemStats)

	mux.HandleFunc("GET "+routeRepositoryState, s.repositoryState)
	mux.HandleFunc("POST "+routeRepositoryOpen, s.repositoryOpen)
	mux.HandleFunc("POST "+routeRepositoryRefresh, s.repositoryRefresh)
	mux.HandleFunc("GET "+routeRepositoryCommit, s.repositoryCommit)
	mux.HandleFunc("GET "+routeRepositoryLog, s.repositoryLog)
	mux.HandleFunc("GET "+routeRepositoryFile, s.repositoryFile)
	mux.HandleFunc("GET "+routeRepositoryFileRange, s.repositoryFileRange)
	mux.HandleFunc("GET "+routeRepositoryFileRaw, s.repositoryFileRaw)
	mux.HandleFunc("GET "+routeRepositoryPullReqs, s.repositoryPullRequests)
	mux.HandleFunc("GET "+routeRepositoryPullReq, s.repositoryPullRequest)
	mux.HandleFunc("POST "+routeWalkthrough, s.walkthroughGenerate)

	mux.HandleFunc("GET "+routeRepositoryBranches, s.repositoryBranches)
	mux.HandleFunc("POST "+routeRepositoryCheckout, s.repositoryCheckout)
	mux.HandleFunc("POST "+routeRepositoryBranchesCreate, s.repositoryCreateBranch)
	mux.HandleFunc("POST "+routeRepositoryBranchesLock, s.lockBranch)
	mux.HandleFunc("POST "+routeRepositoryBranchesUnlock, s.unlockBranch)
	mux.HandleFunc("DELETE "+routeRepositoryBranches, s.deleteBranch)

	mux.HandleFunc("GET "+routeRepositories, s.listRepositories)
	mux.HandleFunc("GET "+routeRepositoriesSearchFiles, s.searchRepositoryFiles)
	mux.HandleFunc("POST "+routeRepositories, s.upsertRepository)
	mux.HandleFunc("DELETE "+routeRepositories, s.removeRepository)
	mux.HandleFunc("GET "+routeRepositoriesCollaborator, s.listCollaborators)
	mux.HandleFunc("POST "+routeRepositoriesCollaborator, s.addCollaborator)
	mux.HandleFunc("DELETE "+routeRepositoriesCollaborator, s.removeCollaborator)

	mux.HandleFunc("POST "+routeReviews, s.createReview)
	mux.HandleFunc("GET "+routeReviews, s.listReviews)
	mux.HandleFunc("GET "+routeReviewDetail, s.reviewDetail)
	mux.HandleFunc("POST "+routeReviewEvents, s.addReviewEvent)
	mux.HandleFunc("POST "+routeReviewComments, s.createReviewComment)
	mux.HandleFunc("PATCH "+routeReviewComment, s.updateReviewComment)
	mux.HandleFunc("PATCH "+routeReviewCommentResolve, s.setReviewCommentResolved)
	mux.HandleFunc("DELETE "+routeReviewComment, s.deleteReviewComment)

	mux.HandleFunc("GET "+routePendingComments, s.listPendingComments)
	mux.HandleFunc("POST "+routePendingComments, s.createPendingComment)
	mux.HandleFunc("PATCH "+routePendingComment, s.updatePendingComment)
	mux.HandleFunc("DELETE "+routePendingComment, s.deletePendingComment)
	mux.HandleFunc("POST "+routePendingCommentsPromote, s.promotePendingComments)

	mux.HandleFunc("GET "+routePreferences, s.getPreferences)
	mux.HandleFunc("POST "+routePreferences, s.savePreferences)

	mux.HandleFunc("GET "+routeUserConfig, s.userConfigGet)
	mux.HandleFunc("GET "+routeUserConfigStream, s.userConfigStream)

	mux.HandleFunc("GET "+routeAuthState, s.authState)
	mux.HandleFunc("POST "+routeAuthSetup, s.authSetup)
	mux.HandleFunc("POST "+routeAuthLogin, s.authLogin)
	mux.HandleFunc("POST "+routeAuthResume, s.authResume)
	mux.HandleFunc("POST "+routeAuthLogout, s.authLogout)
	mux.HandleFunc("POST "+routeAuthWipe, s.authWipe)

	return mux
}
