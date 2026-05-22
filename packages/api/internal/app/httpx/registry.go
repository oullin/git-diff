package httpx

import (
	"github.com/gocanto/git-diff/internal/ai"
	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/userconfig"
)

// services bundles the application services consumed by HTTP handlers.
// Wired in bootstrap.go from the storage repositories.
type services struct {
	auth         *service.AuthService
	reviews      *service.ReviewService
	pending      *service.PendingCommentService
	repos        *service.RepositoryService
	preferences  *service.PreferenceService
	branches     *service.BranchService
	walkthroughs *service.WalkthroughService
}

func newServices(store *storage.Store, providers *ai.Registry, userCfg userconfig.Reader) *services {
	return &services{
		auth: service.NewAuthService(store.Users, store.Sessions, service.AuthConfig{
			BcryptCost:        bcryptCost,
			SessionTTL:        sessionTTL,
			MinPasswordLength: minPasswordLength,
		}),
		reviews:      service.NewReviewService(store.Reviews, store.Comments),
		pending:      service.NewPendingCommentService(store.PendingComments),
		repos:        service.NewRepositoryService(store.Repos),
		preferences:  service.NewPreferenceService(store.Preferences),
		branches:     service.NewBranchService(store.Branches),
		walkthroughs: service.NewWalkthroughService(store.Walkthroughs, providers, userCfg),
	}
}
