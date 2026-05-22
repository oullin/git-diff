package httpx

import (
	"github.com/gocanto/git-diff/internal/ai"
	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
	"github.com/gocanto/git-diff/internal/userconfig"
)

// services is the concrete ServiceRegistry backed by a real *storage.Store.
// Wired in bootstrap.go; handlers consume the ServiceRegistry interface so
// they never see the storage layer.
type services struct {
	auth         *service.AuthService
	reviews      *service.ReviewService
	pending      *service.PendingCommentService
	repos        *service.RepositoryService
	preferences  *service.PreferenceService
	branches     *service.BranchService
	walkthroughs *service.WalkthroughService
}

func newServiceRegistry(store *storage.Store, providers *ai.Registry, userCfg userconfig.Reader) *services {
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

func (s *services) Auth() *service.AuthService                      { return s.auth }
func (s *services) Reviews() *service.ReviewService                 { return s.reviews }
func (s *services) PendingComments() *service.PendingCommentService { return s.pending }
func (s *services) Repositories() *service.RepositoryService        { return s.repos }
func (s *services) Preferences() *service.PreferenceService         { return s.preferences }
func (s *services) Branches() *service.BranchService                { return s.branches }
func (s *services) Walkthroughs() *service.WalkthroughService       { return s.walkthroughs }
