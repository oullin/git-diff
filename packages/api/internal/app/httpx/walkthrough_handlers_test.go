package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWalkthroughRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/walkthrough", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestWalkthroughCommitKindRequiresSHA(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]any{"path": srv.Repo, "kind": "commit"})
	resp, err := http.Post(ts.URL+"/v1/walkthrough", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for commit kind without sha, got %d", resp.StatusCode)
	}
}

func TestWalkthroughBadPathReturnsBadRequest(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	// Use a clearly invalid path; ReadRepositoryState shells out to git and
	// will error out, which the handler maps to 400.
	body, _ := json.Marshal(map[string]any{"path": "/tmp/does-not-exist-walkthrough", "kind": "working"})
	resp, err := http.Post(ts.URL+"/v1/walkthrough", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
