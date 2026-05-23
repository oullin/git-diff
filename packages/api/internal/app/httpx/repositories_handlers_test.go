package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestListRepositoriesAuthRequired(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repositories")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUpsertAndListRepositories(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"path": "/my/repo", "name": "myrepo"})
	resp, err := http.Post(ts.URL+"/v1/repositories", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upsert: expected 200, got %d", resp.StatusCode)
	}

	listResp, err := http.Get(ts.URL + "/v1/repositories")

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	defer listResp.Body.Close()

	var payload struct {
		Repositories []struct {
			Path string `json:"path"`
			Name string `json:"name"`
		} `json:"repositories"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(payload.Repositories) != 1 || payload.Repositories[0].Path != "/my/repo" {
		t.Fatalf("unexpected list: %+v", payload)
	}
}

func TestUpsertRepositoryRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/repositories", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRemoveRepositoryReturnsBadRequestWhenPathMissing(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/repositories", nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRemoveRepositoryReturns204OnSuccess(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"path": "/my/repo", "name": "r"})

	if resp, err := http.Post(ts.URL+"/v1/repositories", "application/json", bytes.NewReader(body)); err != nil {
		t.Fatalf("post: %v", err)
	} else {
		resp.Body.Close()
	}

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/repositories?path=/my/repo", nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestListCollaboratorsRequiresPath(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/repositories/collaborators")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAddAndListCollaborator(t *testing.T) {
	srv, store := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	// Seed a second user to be added as a collaborator.
	other, err := store.Users.EnsureUser(t.Context(), "bob")

	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Create the repo first via the API so it belongs to the authenticated user.
	body, _ := json.Marshal(map[string]string{"path": "/repo", "name": "r"})

	if resp, err := http.Post(ts.URL+"/v1/repositories", "application/json", bytes.NewReader(body)); err != nil {
		t.Fatalf("upsert: %v", err)
	} else {
		resp.Body.Close()
	}

	addBody, _ := json.Marshal(map[string]any{
		"path":   "/repo",
		"userId": other.ID,
		"role":   "write",
	})
	resp, err := http.Post(ts.URL+"/v1/repositories/collaborators", "application/json", bytes.NewReader(addBody))

	if err != nil {
		t.Fatalf("add: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add: expected 201, got %d", resp.StatusCode)
	}

	listResp, err := http.Get(ts.URL + "/v1/repositories/collaborators?path=/repo")

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	defer listResp.Body.Close()

	var payload struct {
		Collaborators []map[string]any `json:"collaborators"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(payload.Collaborators) != 1 {
		t.Fatalf("expected 1 collaborator, got %d: %v", len(payload.Collaborators), payload)
	}
}

func TestRemoveCollaboratorRequiresIntegerUserID(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/repositories/collaborators?path=/r&userId=notanint", nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRemoveCollaboratorRequiresQueryParams(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	// Missing path.
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/repositories/collaborators?userId=1", nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRemoveCollaboratorReturns204OnSuccess(t *testing.T) {
	srv, store := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	other, _ := store.Users.EnsureUser(t.Context(), "bob")

	body, _ := json.Marshal(map[string]string{"path": "/repo", "name": "r"})

	if resp, err := http.Post(ts.URL+"/v1/repositories", "application/json", bytes.NewReader(body)); err != nil {
		t.Fatalf("upsert: %v", err)
	} else {
		resp.Body.Close()
	}

	add, _ := json.Marshal(map[string]any{"path": "/repo", "userId": other.ID, "role": "read"})

	if resp, err := http.Post(ts.URL+"/v1/repositories/collaborators", "application/json", bytes.NewReader(add)); err != nil {
		t.Fatalf("add: %v", err)
	} else {
		resp.Body.Close()
	}

	url := fmt.Sprintf("%s/v1/repositories/collaborators?path=/repo&userId=%d", ts.URL, other.ID)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}
