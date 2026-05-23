package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func postReview(t *testing.T, url string, body any) *http.Response {
	t.Helper()

	raw, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	return resp
}

func TestCreateReviewRequiresAuth(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp := postReview(t, ts.URL+"/v1/reviews", map[string]string{"repoRoot": "/r"})
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateReviewRoundTrip(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp := postReview(t, ts.URL+"/v1/reviews", map[string]any{
		"repoRoot": "/r",
		"branch":   "main",
		"headSha":  "abc",
	})

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var review struct {
		ID int64 `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&review); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if review.ID == 0 {
		t.Fatalf("expected an id, got %+v", review)
	}
}

func TestCreateReviewRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/reviews", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListReviewsRejectsInvalidLimit(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/reviews?limit=abc")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListReviewsReturnsCreated(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp := postReview(t, ts.URL+"/v1/reviews", map[string]string{"repoRoot": "/r"})
	resp.Body.Close()

	listResp, err := http.Get(ts.URL + "/v1/reviews")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer listResp.Body.Close()

	var payload struct {
		Reviews []map[string]any `json:"reviews"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(payload.Reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(payload.Reviews))
	}
}

func TestReviewDetailReturnsBadRequestForBadID(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/reviews/notanint")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReviewDetailReturnsNotFoundForMissingReview(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/reviews/9999")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestReviewFullLifecycle(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	create := postReview(t, ts.URL+"/v1/reviews", map[string]string{"repoRoot": "/r"})

	defer create.Body.Close()

	var created struct {
		ID int64 `json:"id"`
	}

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Event.
	eventResp := postReview(t, fmt.Sprintf("%s/v1/reviews/%d/events", ts.URL, created.ID), map[string]string{"type": "viewed"})

	if eventResp.StatusCode != http.StatusCreated {
		t.Fatalf("event: expected 201, got %d", eventResp.StatusCode)
	}

	eventResp.Body.Close()

	// Comment.
	commentResp := postReview(t, fmt.Sprintf("%s/v1/reviews/%d/comments", ts.URL, created.ID), map[string]any{
		"filePath": "a.go",
		"side":     "right",
		"line":     1,
		"bodyHtml": "<p>v1</p>",
	})

	if commentResp.StatusCode != http.StatusCreated {
		t.Fatalf("comment: expected 201, got %d", commentResp.StatusCode)
	}

	var comment struct {
		ID int64 `json:"id"`
	}

	if err := json.NewDecoder(commentResp.Body).Decode(&comment); err != nil {
		t.Fatalf("decode comment: %v", err)
	}

	commentResp.Body.Close()

	// PATCH the comment.
	patchBody, _ := json.Marshal(map[string]string{"bodyHtml": "<p>v2</p>"})
	patchReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/v1/reviews/%d/comments/%d", ts.URL, created.ID, comment.ID), bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)

	if err != nil {
		t.Fatalf("patch: %v", err)
	}

	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("patch: expected 200, got %d", patchResp.StatusCode)
	}

	patchResp.Body.Close()

	// DELETE the comment.
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/reviews/%d/comments/%d", ts.URL, created.ID, comment.ID), nil)
	delResp, err := http.DefaultClient.Do(delReq)

	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", delResp.StatusCode)
	}
}

func TestUpdateReviewCommentRejectsBadIDs(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"bodyHtml": "<p>x</p>"})
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/v1/reviews/abc/comments/1", bytes.NewReader(body))
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
