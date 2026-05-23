package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestListPendingCommentsAuthRequired(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/pending-comments")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreatePendingCommentRejectsMissingRequiredFields(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"bodyHtml": "<p>x</p>"})
	resp, err := http.Post(ts.URL+"/v1/pending-comments", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 from invalid input, got %d", resp.StatusCode)
	}
}

func TestCreatePendingCommentRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/pending-comments", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPendingCommentLifecycle(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	create, _ := json.Marshal(map[string]any{
		"repoRoot":    "/r",
		"contextKind": "branch",
		"contextSha":  "sha",
		"filePath":    "a.go",
		"diffSection": "diff-1",
		"side":        "right",
		"lineNumber":  1,
		"bodyHtml":    "<p>v1</p>",
	})

	resp, err := http.Post(ts.URL+"/v1/pending-comments", "application/json", bytes.NewReader(create))

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", resp.StatusCode)
	}

	var pending struct {
		ID int64 `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pending); err != nil {
		t.Fatalf("decode: %v", err)
	}

	resp.Body.Close()

	// List by query scope.
	listResp, err := http.Get(ts.URL + "/v1/pending-comments?path=/r&kind=branch&sha=sha")

	if err != nil {
		t.Fatalf("list: %v", err)
	}

	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", listResp.StatusCode)
	}

	// PATCH.
	patchBody, _ := json.Marshal(map[string]string{"bodyHtml": "<p>v2</p>"})
	patchReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/v1/pending-comments/%d", ts.URL, pending.ID), bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)

	if err != nil {
		t.Fatalf("patch: %v", err)
	}

	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("patch: expected 200, got %d", patchResp.StatusCode)
	}

	patchResp.Body.Close()

	// DELETE.
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/pending-comments/%d", ts.URL, pending.ID), nil)
	delResp, err := http.DefaultClient.Do(delReq)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", delResp.StatusCode)
	}
}

func TestPromotePendingCommentsRequiresReviewID(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]int64{"reviewId": 0})
	resp, err := http.Post(ts.URL+"/v1/pending-comments/promote", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("promote: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdatePendingCommentRejectsBadID(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"bodyHtml": "<p>x</p>"})
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/v1/pending-comments/abc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("patch: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
