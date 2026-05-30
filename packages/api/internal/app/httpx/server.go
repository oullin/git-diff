package httpx

import (
	"io"
	"net/http"

	"github.com/oullin/git-diff/internal/app/setting"
	"github.com/oullin/git-diff/internal/service"
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

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
