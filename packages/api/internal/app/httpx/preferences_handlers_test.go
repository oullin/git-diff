package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGetPreferencesAuthRequired(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/preferences")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestGetPreferencesReturnsEmptyForNewUser(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/preferences")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var prefs struct {
		Values map[string]string `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if prefs.Values == nil {
		t.Fatalf("expected non-nil values map")
	}
}

func TestSavePreferencesRejectsInvalidJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/preferences", "application/json", strings.NewReader("not-json"))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad JSON, got %d", resp.StatusCode)
	}
}

func TestSavePreferencesRoundTrip(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{Authenticated: true})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]any{"values": map[string]string{"theme": "dark"}})
	resp, err := http.Post(ts.URL+"/v1/preferences", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("post: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var prefs struct {
		Values map[string]string `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if prefs.Values["theme"] != "dark" {
		t.Fatalf("expected theme=dark in response, got %#v", prefs.Values)
	}
}
