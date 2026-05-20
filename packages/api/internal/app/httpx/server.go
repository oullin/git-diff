package httpx

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/storage"
)

type StoreFactory func(context.Context) (*storage.Store, func(), error)

type ServeConfig struct {
	Home   string
	Repo   string
	Stderr io.Writer
}

type Server struct {
	Home     string
	Repo     string
	Settings setting.RuntimeSettings
	Store    StoreFactory
	Auth     *AuthState
}

func Serve(args []string, cfg ServeConfig) int {
	fs := flag.NewFlagSet("serve-http", flag.ContinueOnError)
	fs.SetOutput(cfg.Stderr)

	socketPath := fs.String("socket", "", "Unix socket path")
	repoRoot := fs.String("repo-root", "", "Repository root")
	dbPath := fs.String("db", "", "Review SQLite database path")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *socketPath == "" {
		fmt.Fprintln(cfg.Stderr, "missing --socket")

		return 2
	}

	settings := setting.RuntimeSettings{
		RepoRoot:     *repoRoot,
		DatabasePath: *dbPath,
	}
	validation := setting.ValidateRuntimeSettings(cfg.Home, cfg.Repo, settings)

	if !validation.Valid {
		for _, check := range validation.Checks {
			if check.Status == setting.CheckError {
				fmt.Fprintf(cfg.Stderr, "invalid settings: %s: %s\n", check.Label, check.Message)
			}
		}

		return 2
	}

	settings = validation.Settings
	repo := settings.RepoRoot

	store, err := storage.Open(context.Background(), settings.DatabasePath)

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "open review database: %v\n", err)

		return 1
	}

	defer func() {
		if err := store.Close(); err != nil {
			fmt.Fprintf(cfg.Stderr, "close review database: %v\n", err)
		}
	}()

	if err := os.Remove(*socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(cfg.Stderr, "remove stale http socket: %v\n", err)

		return 1
	}

	listener, err := net.Listen("unix", *socketPath)

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "listen on http socket: %v\n", err)

		return 1
	}

	defer func() {
		if err := listener.Close(); err != nil {
			fmt.Fprintf(cfg.Stderr, "close http listener: %v\n", err)
		}

		if err := os.Remove(*socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(cfg.Stderr, "remove http socket: %v\n", err)
		}
	}()

	osUsername := resolveOSUsername()

	if _, err := store.EnsureUser(context.Background(), osUsername); err != nil {
		fmt.Fprintf(cfg.Stderr, "seed os user %q: %v\n", osUsername, err)

		return 1
	}

	authState := NewAuthState(osUsername)

	appServer := Server{
		Home:     cfg.Home,
		Repo:     repo,
		Settings: settings,
		Store: func(context.Context) (*storage.Store, func(), error) {
			return store, func() {}, nil
		},
		Auth: authState,
	}
	server := &http.Server{Handler: NewServerHandler(ServerHandlerConfig{
		Mux:           appServer.requireAuth(appServer.BuildMux()),
		SafeQueryKeys: []string{"limit"},
	})}

	if err := RunServer(*socketPath, listener, server); err != nil {
		fmt.Fprintf(cfg.Stderr, "serve http: %v\n", err)

		return 1
	}

	return 0
}

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
	mux.HandleFunc("GET /v1/settings", s.getSettings)
	mux.HandleFunc("POST /v1/settings/validate", s.validateSettings)
	mux.HandleFunc("GET /v1/preferences", s.getPreferences)
	mux.HandleFunc("POST /v1/preferences", s.savePreferences)
	mux.HandleFunc("GET /v1/auth/state", s.authState)
	mux.HandleFunc("POST /v1/auth/setup", s.authSetup)
	mux.HandleFunc("POST /v1/auth/login", s.authLogin)
	mux.HandleFunc("POST /v1/auth/resume", s.authResume)
	mux.HandleFunc("POST /v1/auth/logout", s.authLogout)
	mux.HandleFunc("POST /v1/auth/wipe", s.authWipe)

	return mux
}

func (s Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func resolveOSUsername() string {
	if u, err := user.Current(); err == nil {
		if name := strings.TrimSpace(u.Username); name != "" {
			return name
		}
	}

	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name
	}

	if name := strings.TrimSpace(os.Getenv("USERNAME")); name != "" {
		return name
	}

	return "user"
}
