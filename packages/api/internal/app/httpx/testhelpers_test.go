package httpx

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/oullin/git-diff/internal/ai"
	"github.com/oullin/git-diff/internal/app/setting"
	"github.com/oullin/git-diff/internal/service"
	"github.com/oullin/git-diff/internal/storage"
	"github.com/oullin/git-diff/internal/usercfg"
)

// testServerOptions tweaks the test server before BuildMux is called. Most
// tests use the defaults; set Authenticated=true (or supply a seeded user)
// when the requireAuth middleware needs to be satisfied.
type testServerOptions struct {
	// Authenticated, when true, seeds an OS user, marks the session as
	// signed in, and exposes its ID on the returned Server. The default OS
	// username is "test" — override via OSUsername.
	Authenticated bool
	OSUsername    string

	// Repo overrides Server.Repo (otherwise t.TempDir()).
	Repo string

	// Home overrides Server.Home (otherwise t.TempDir()).
	Home string

	// ExtraProviders is appended to the ai.Registry after the stub
	// "anthropic" provider is registered.
	ExtraProviders []ai.Provider
}

// testServer wires a Server with a real *storage.Store, all services, a
// stub usercfg.Reader/Broker, and an ai.Registry containing a stub provider
// registered as the default "anthropic". The returned Server matches the
// production bootstrap shape; tests that need to call BuildMux() should do
// so on the returned value.
func testServer(t *testing.T, opts testServerOptions) (Server, *storage.Store) {
	t.Helper()

	home := opts.Home

	if home == "" {
		home = t.TempDir()
	}

	repo := opts.Repo

	if repo == "" {
		repo = t.TempDir()
	}

	osUsername := opts.OSUsername

	if osUsername == "" {
		osUsername = "test"
	}

	store, err := storage.Open(context.Background(), filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	user, err := store.Users.EnsureUser(context.Background(), osUsername)

	if err != nil {
		t.Fatalf("ensure os user: %v", err)
	}

	reader := usercfg.NewAtomicReader(usercfg.Defaults())
	broker := usercfg.NewBroker()

	providers := ai.NewRegistry()
	providers.Register(newStubProvider("anthropic"))

	for _, p := range opts.ExtraProviders {
		providers.Register(p)
	}

	state := NewAuthState(osUsername)

	if opts.Authenticated {
		session, err := store.Sessions.CreateSession(context.Background(), user.ID, sessionTTL)

		if err != nil {
			t.Fatalf("create session: %v", err)
		}

		state.Set(user.ID, session.RawToken)
	}

	srv := Server{
		Home:             home,
		Repo:             repo,
		Settings:         setting.DefaultRuntimeSettings(home, repo),
		Session:          state,
		UserConfig:       reader,
		UserConfigEvents: broker,
		UserConfigPath:   usercfg.DefaultPath(home),

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
		walkthroughs:  service.NewWalkthroughService(store.Walkthroughs, providers, reader),
	}

	return srv, store
}

// startHTTPServer spins up an httptest.Server backed by the supplied
// Server's BuildMux. The unauthenticated form skips requireAuth; pass
// withAuth=true to exercise the same middleware production uses.
func startHTTPServer(t *testing.T, srv Server, withAuth bool) *httptest.Server {
	t.Helper()

	handler := srv.BuildMux()

	if withAuth {
		ts := httptest.NewServer(srv.requireAuth(handler))

		t.Cleanup(ts.Close)

		return ts
	}

	ts := httptest.NewServer(handler)

	t.Cleanup(ts.Close)

	return ts
}
