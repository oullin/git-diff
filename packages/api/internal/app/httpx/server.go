package httpx

import (
	"io"
	"net/http"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/userconfig"
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
	Auth             *AuthState
	Services         *services
	UserConfig       userconfig.Reader
	UserConfigEvents *userconfig.Broker
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
