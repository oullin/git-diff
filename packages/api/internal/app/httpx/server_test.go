package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/storage"
)

func TestHTTPHealthz(t *testing.T) {
	server := httptest.NewServer(testHTTPServer(t, "/Users/gus", "/repo").BuildMux())

	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/healthz")

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func testHTTPServer(t *testing.T, home, repo string) Server {
	t.Helper()

	settings := setting.DefaultRuntimeSettings(home, repo)

	return Server{
		Home:     home,
		Repo:     repo,
		Settings: settings,
		Store: func(ctx context.Context) (*storage.Store, func(), error) {
			store, err := storage.Open(ctx, settings.DatabasePath)

			if err != nil {
				return nil, nil, err
			}

			return store, func() { _ = store.Close() }, nil
		},
		Auth: NewAuthState("test"),
	}
}
