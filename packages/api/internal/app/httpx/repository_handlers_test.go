package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRepositoryStateForGitRepo(t *testing.T) {
	repo := newGitRepo(t)

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/state")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRepositoryStateForNonGitPathReturns400(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Repo: t.TempDir()})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/state")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-git path, got %d", resp.StatusCode)
	}
}

func TestRepositoryOpenRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/repository/open", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryOpenForGitRepo(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"path": repo.Root})
	resp, err := http.Post(ts.URL+"/v1/repository/open", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRepositoryCommitRequiresSHA(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/commit")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryCommitForKnownSHA(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	url := fmt.Sprintf("%s/v1/repository/commit?sha=%s", ts.URL, repo.HeadSHA())
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRepositoryLogReturnsCommits(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/log")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Commits []map[string]any `json:"commits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(payload.Commits) == 0 {
		t.Fatalf("expected at least one commit, got %v", payload.Commits)
	}
}

func TestRepositoryLogRejectsBadLimit(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/log?limit=nope")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRequiresPath(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileReadsKnownFile(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file?path=README.md")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRangeRejectsMissingParams(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	cases := []string{
		"/v1/repository/file-range",
		"/v1/repository/file-range?path=README.md",
		"/v1/repository/file-range?path=README.md&startLine=1",
	}

	for _, p := range cases {
		resp, err := http.Get(ts.URL + p)

		if err != nil {
			t.Fatalf("get %s: %v", p, err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", p, resp.StatusCode)
		}
	}
}

func TestRepositoryFileRawReadsWorktree(t *testing.T) {
	repo := newGitRepo(t)
	want := []byte("PNG-WORKTREE-BYTES")
	repo.WriteFile("images/logo.png", string(want))

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=images/logo.png")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected image/png, got %q", ct)
	}

	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("expected no-store, got %q", cc)
	}

	body, _ := io.ReadAll(resp.Body)

	if !bytes.Equal(body, want) {
		t.Fatalf("body mismatch: got %q want %q", body, want)
	}
}

func TestRepositoryFileRawReadsCommittedBlobAtRef(t *testing.T) {
	repo := newGitRepo(t)
	committed := []byte("PNG-AT-HEAD")
	repo.WriteFile("images/logo.png", string(committed))
	repo.Commit("add logo")

	// Overwrite the worktree with different bytes after committing; the ref
	// read must return the committed bytes, not the worktree bytes.
	repo.WriteFile("images/logo.png", "WORKTREE-DIFFERENT")

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=images/logo.png&ref=HEAD")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	if !bytes.Equal(body, committed) {
		t.Fatalf("expected committed bytes %q, got %q", committed, body)
	}
}

func TestRepositoryFileRawReadsStagedBlobAtIndex(t *testing.T) {
	repo := newGitRepo(t)
	repo.WriteFile("images/logo.png", "OLD")
	repo.Commit("add logo")

	staged := []byte("STAGED-VERSION")
	repo.WriteFile("images/logo.png", string(staged))
	repo.run("-C", repo.Root, "add", "images/logo.png")

	// Make worktree differ from index again, so a `:0` read can be
	// distinguished from a worktree read.
	repo.WriteFile("images/logo.png", "WORKTREE-AGAIN")

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=images/logo.png&ref=:0")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	if !bytes.Equal(body, staged) {
		t.Fatalf("expected staged bytes %q, got %q", staged, body)
	}
}

func TestRepositoryFileRawRequiresPath(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRawReturns404ForMissingWorktreeFile(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=does/not/exist.png")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRawReturns404ForMissingBlobAtRef(t *testing.T) {
	repo := newGitRepo(t)
	// File exists in worktree but was never committed → not present at HEAD.
	repo.WriteFile("new.png", "JUST-IN-WORKTREE")

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=new.png&ref=HEAD")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRawRejectsPathTraversal(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	cases := []string{
		"/v1/repository/file/raw?path=../etc/passwd",
		"/v1/repository/file/raw?path=/etc/passwd",
	}

	for _, p := range cases {
		resp, err := http.Get(ts.URL + p)

		if err != nil {
			t.Fatalf("get %s: %v", p, err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", p, resp.StatusCode)
		}
	}
}

func TestRepositoryFileRawRejectsOversizeWithStatus413(t *testing.T) {
	repo := newGitRepo(t)
	// 3 MB > maxFileReadBytes (2 MB).
	big := bytes.Repeat([]byte{'x'}, 3*1024*1024)
	repo.WriteFile("big.png", string(big))

	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=big.png")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}
}

func TestRepositoryFileRawContentTypeByExtension(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	// Just enough bytes to satisfy a successful read; Content-Type is derived
	// from the extension alone, not from the body.
	for path, wantPrefix := range map[string]string{
		"a.png":  "image/png",
		"a.jpg":  "image/jpeg",
		"a.gif":  "image/gif",
		"a.webp": "image/webp",
		"a.svg":  "image/svg+xml",
	} {
		repo.WriteFile(path, "BYTES")

		resp, err := http.Get(ts.URL + "/v1/repository/file/raw?path=" + path)

		if err != nil {
			t.Fatalf("%s: get: %v", path, err)
		}

		ct := resp.Header.Get("Content-Type")
		resp.Body.Close()

		if !strings.HasPrefix(ct, wantPrefix) {
			t.Fatalf("%s: expected Content-Type starting with %q, got %q", path, wantPrefix, ct)
		}
	}
}

func TestRepositoryFileRangeRejectsBadInteger(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/file-range?path=README.md&startLine=a&endLine=1")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryBranchesReturnsMain(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repository/branches")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Branches []string `json:"branches"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}

	found := false

	for _, name := range payload.Branches {
		if name == "main" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected main in branches, got %v", payload.Branches)
	}
}

func TestRepositoryCreateAndDeleteBranch(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Authenticated: true, Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	// Create branch via API.
	body, _ := json.Marshal(map[string]string{"name": "feature/x"})
	resp, err := http.Post(ts.URL+"/v1/repository/branches/create", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create: expected 200, got %d", resp.StatusCode)
	}

	// Branch creation also checks it out, so deleting feature/x must first
	// switch back to main.
	repo.Checkout("main", false)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/repository/branches?name=feature/x", nil)
	delResp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", delResp.StatusCode)
	}
}

func TestRepositoryCheckoutRejectsBadJSON(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Authenticated: true, Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/repository/checkout", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRepositoryCheckoutSucceedsOnExistingBranch(t *testing.T) {
	repo := newGitRepo(t)
	repo.Branch("feature/y")
	srv, _ := testServer(t, testServerOptions{Authenticated: true, Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"branch": "feature/y"})
	resp, err := http.Post(ts.URL+"/v1/repository/checkout", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLockBranchRequiresRegisteredRepo(t *testing.T) {
	repo := newGitRepo(t)
	srv, _ := testServer(t, testServerOptions{Authenticated: true, Repo: repo.Root})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"name": "main"})
	resp, err := http.Post(ts.URL+"/v1/repository/branches/lock", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	// Repository is a real git tree but not registered in the DB; the service
	// returns ErrRepositoryNotFound, which the handler maps to 400 via the
	// branch error switch's default branch.
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unregistered repo, got %d", resp.StatusCode)
	}
}
