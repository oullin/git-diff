package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/storage"
)

func TestHTTPSettingsValidation(t *testing.T) {
	home := t.TempDir()
	repo := writeSettingsRepo(t)
	server := httptest.NewServer(testHTTPServer(t, home, repo).BuildMux())

	defer server.Close()

	get := getJSON(t, server.URL+"/v1/settings")

	if valid, _ := get["valid"].(bool); !valid {
		t.Fatalf("settings response = %#v", get)
	}

	settings, _ := get["settings"].(map[string]any)

	if settings["repoRoot"] != repo {
		t.Fatalf("repoRoot = %#v", settings["repoRoot"])
	}

	body := bytes.NewBufferString(`{"settings":{"repoRoot":"` + filepath.Join(home, "missing") + `"}}`)
	resp, err := http.Post(server.URL+"/v1/settings/validate", "application/json", body)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	var validation map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		t.Fatal(err)
	}

	if valid, _ := validation["valid"].(bool); valid {
		t.Fatalf("expected invalid settings, got %#v", validation)
	}
}

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

func writeSettingsRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repo

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

	return repo
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()

	resp, err := http.Get(url)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		t.Fatalf("GET %s = %d: %s", url, resp.StatusCode, body)
	}

	var body map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	return body
}

var _ = os.PathSeparator
