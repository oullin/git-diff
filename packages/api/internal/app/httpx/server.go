package httpx

import (
	"context"
	"io"
	"net/http"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/service"
	"github.com/gocanto/git-diff/internal/storage"
)

type StoreFactory func(context.Context) (*storage.Store, func(), error)

type ServeConfig struct {
	Home   string
	Repo   string
	Stderr io.Writer
}

type Server struct {
	Home          string
	Repo          string
	Settings      setting.RuntimeSettings
	Store         StoreFactory
	Auth          *AuthState
	AuthService   *service.AuthService
	ReviewService *service.ReviewService
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
