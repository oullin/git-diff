package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocanto/git-diff/internal/app/setting"
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

	return Server{
		Home:     home,
		Repo:     repo,
		Settings: setting.DefaultRuntimeSettings(home, repo),
		Auth:     NewAuthState("test"),
	}
}
