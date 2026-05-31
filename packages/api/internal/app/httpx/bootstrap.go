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

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/app/setting"
	"github.com/oullin/git-diff/internal/storage"
	"github.com/oullin/git-diff/internal/usercfg"
)

// Serve is the CLI entry point for the "serve-http" subcommand.
func Serve(args []string, cfg ServeConfig) int {
	settings, socketPath, exit := parseServeFlags(args, cfg)

	if exit != 0 {
		return exit
	}

	store, exit := openStore(settings.DatabasePath, cfg.Stderr)

	if exit != 0 {
		return exit
	}

	defer closeStore(store, cfg.Stderr)

	listener, cleanup, exit := openSocket(socketPath, cfg.Stderr)

	if exit != 0 {
		return exit
	}

	defer cleanup()

	osUsername := resolveOSUsername()

	if _, err := store.Users.EnsureUser(context.Background(), osUsername); err != nil {
		fmt.Fprintf(cfg.Stderr, "seed os user %q: %v\n", osUsername, err)

		return 1
	}

	userCfg, err := usercfg.NewService(cfg.Home)

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "user config: %v\n", err)

		return 1
	}

	watchCtx, cancelWatch := context.WithCancel(context.Background())

	defer cancelWatch()

	go func() {
		_ = userCfg.Run(watchCtx)
	}()

	providers := ai.NewRegistry()
	providers.Register(ai.NewAnthropicProvider(userCfg.Reader))
	providers.Register(ai.NewCodexProvider(userCfg.Reader))

	appServer := newServer(cfg, settings, store, providers, userCfg, osUsername)

	server := &http.Server{Handler: NewServerHandler(ServerHandlerConfig{
		Mux:           withRequestCaches(appServer.requireAuth(appServer.BuildMux())),
		SafeQueryKeys: []string{"limit"},
	})}

	if err := RunServer(socketPath, listener, server); err != nil {
		fmt.Fprintf(cfg.Stderr, "serve http: %v\n", err)

		return 1
	}

	return 0
}

func parseServeFlags(args []string, cfg ServeConfig) (setting.RuntimeSettings, string, int) {
	fs := flag.NewFlagSet("serve-http", flag.ContinueOnError)
	fs.SetOutput(cfg.Stderr)

	socketPath := fs.String("socket", "", "Unix socket path")
	repoRoot := fs.String("repo-root", "", "Repository root")
	dbPath := fs.String("db", "", "Review SQLite database path")

	if err := fs.Parse(args); err != nil {
		return setting.RuntimeSettings{}, "", 2
	}

	if *socketPath == "" {
		fmt.Fprintln(cfg.Stderr, "missing --socket")

		return setting.RuntimeSettings{}, "", 2
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

		return setting.RuntimeSettings{}, "", 2
	}

	return validation.Settings, *socketPath, 0
}

func openStore(path string, stderr io.Writer) (*storage.Store, int) {
	store, err := storage.Open(context.Background(), path)

	if err != nil {
		fmt.Fprintf(stderr, "open review database: %v\n", err)

		return nil, 1
	}

	return store, 0
}

func closeStore(store *storage.Store, stderr io.Writer) {
	if err := store.Close(); err != nil {
		fmt.Fprintf(stderr, "close review database: %v\n", err)
	}
}

func openSocket(socketPath string, stderr io.Writer) (net.Listener, func(), int) {
	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(stderr, "remove stale http socket: %v\n", err)

		return nil, func() {}, 1
	}

	listener, err := net.Listen("unix", socketPath)

	if err != nil {
		fmt.Fprintf(stderr, "listen on http socket: %v\n", err)

		return nil, func() {}, 1
	}

	cleanup := func() {
		if err := listener.Close(); err != nil {
			fmt.Fprintf(stderr, "close http listener: %v\n", err)
		}

		if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(stderr, "remove http socket: %v\n", err)
		}
	}

	return listener, cleanup, 0
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
