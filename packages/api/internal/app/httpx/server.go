package httpx

import (
	"io"
	"net/http"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/app/setting"
	"github.com/oullin/git-diff/internal/service"
	"github.com/oullin/git-diff/internal/storage"
	"github.com/oullin/git-diff/internal/usercfg"
)

type ServeConfig struct {
	Home   string
	Repo   string
	Stderr io.Writer
}

type Server struct {
	Home             string
	Repo             string
	Settings         setting.RuntimeSettings
	Session          *AuthState
	UserConfig       usercfg.Reader
	UserConfigEvents *usercfg.Broker
	UserConfigPath   string

	auth          *service.AuthService
	reviews       *service.ReviewService
	pending       *service.PendingCommentService
	repos         *service.RepositoryService
	collaborators *service.CollaboratorService
	preferences   *service.PreferenceService
	branches      *service.BranchService
	walkthroughs  *service.WalkthroughService
}

// newServer is the composition root for the HTTP server's service graph: it
// wires the storage repositories, AI providers, and user config into the
// application services. Keeping this here leaves Serve() as pure orchestration
// (parse flags, open store/socket, run server).
func newServer(
	cfg ServeConfig,
	settings setting.RuntimeSettings,
	store *storage.Store,
	providers *ai.Registry,
	userCfg *usercfg.Service,
	osUsername string,
) Server {
	return Server{
		Home:             cfg.Home,
		Repo:             settings.RepoRoot,
		Settings:         settings,
		Session:          NewAuthState(osUsername),
		UserConfig:       userCfg.Reader,
		UserConfigEvents: userCfg.Broker,
		UserConfigPath:   userCfg.Path,

		auth: service.NewAuthService(store.Users, store.Sessions, service.AuthConfig{
			BcryptCost:        bcryptCost,
			SessionTTL:        sessionTTL,
			MinPasswordLength: minPasswordLength,
		}),
		reviews:       service.NewReviewService(store.Reviews, store.ReviewEvents, store.Comments),
		pending:       service.NewPendingCommentService(store.PendingComments),
		repos:         service.NewRepositoryService(store.Repos),
		collaborators: service.NewCollaboratorService(store.Collaborators),
		preferences:   service.NewPreferenceService(store.Preferences),
		branches:      service.NewBranchService(store.Branches, store.Repos),
		walkthroughs:  service.NewWalkthroughService(store.Walkthroughs, providers, userCfg.Reader),
	}
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
