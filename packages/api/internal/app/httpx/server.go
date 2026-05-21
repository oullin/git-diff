package httpx

import (
	"io"
	"net/http"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/service"
)

type ServeConfig struct {
	Home   string
	Repo   string
	Stderr io.Writer
}

// ServiceRegistry hides the concrete service-construction details from the
// HTTP layer. Handlers depend on this interface; bootstrap.go provides a
// concrete implementation backed by the storage repositories.
type ServiceRegistry interface {
	Auth() *service.AuthService
	Reviews() *service.ReviewService
	PendingComments() *service.PendingCommentService
	Repositories() *service.RepositoryService
	Preferences() *service.PreferenceService
	Branches() *service.BranchService
	Walkthroughs() *service.WalkthroughService
}

type Server struct {
	Home     string
	Repo     string
	Settings setting.RuntimeSettings
	Auth     *AuthState
	Services ServiceRegistry
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
