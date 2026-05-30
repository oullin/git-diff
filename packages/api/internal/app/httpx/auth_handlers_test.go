package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAuthStateReportsNeedsSetupForFreshUser(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{OSUsername: "fresh"})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/auth/state")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var state struct {
		NeedsSetup      bool `json:"needsSetup"`
		IsAuthenticated bool `json:"isAuthenticated"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if !state.NeedsSetup {
		t.Fatalf("expected NeedsSetup=true for fresh user")
	}

	if state.IsAuthenticated {
		t.Fatalf("expected IsAuthenticated=false for unauthenticated session")
	}
}

func TestAuthSetupShortPasswordRejected(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"password": "abc"})
	resp, err := http.Post(ts.URL+"/v1/auth/setup", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAuthSetupRejectsBadJSON(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Post(ts.URL+"/v1/auth/setup", "application/json", strings.NewReader("nope"))

	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAuthSetupLoginAndLogoutLifecycle(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	// Setup.
	setupBody, _ := json.Marshal(map[string]string{"password": "secret123"})

	if resp, err := http.Post(ts.URL+"/v1/auth/setup", "application/json", bytes.NewReader(setupBody)); err != nil {
		t.Fatalf("setup: %v", err)
	} else {
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("setup status: %d", resp.StatusCode)
		}

		resp.Body.Close()
	}

	// Setup-again must fail with 409 conflict.
	if resp, err := http.Post(ts.URL+"/v1/auth/setup", "application/json", bytes.NewReader(setupBody)); err != nil {
		t.Fatalf("setup-again: %v", err)
	} else {
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409 on second setup, got %d", resp.StatusCode)
		}

		resp.Body.Close()
	}

	// Login with wrong password -> 401.
	wrong, _ := json.Marshal(map[string]string{"password": "wrong"})

	if resp, err := http.Post(ts.URL+"/v1/auth/login", "application/json", bytes.NewReader(wrong)); err != nil {
		t.Fatalf("login wrong: %v", err)
	} else {
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong password, got %d", resp.StatusCode)
		}

		resp.Body.Close()
	}

	// Login with right password + remember=true -> 200 with token.
	good, _ := json.Marshal(map[string]any{"password": "secret123", "remember": true})

	loginResp, err := http.Post(ts.URL+"/v1/auth/login", "application/json", bytes.NewReader(good))

	if err != nil {
		t.Fatalf("login: %v", err)
	}

	var login struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(loginResp.Body).Decode(&login); err != nil {
		t.Fatalf("decode: %v", err)
	}

	loginResp.Body.Close()

	if login.Token == "" {
		t.Fatalf("expected token on remember=true")
	}

	// Logout returns 204.
	logoutResp, err := http.Post(ts.URL+"/v1/auth/logout", "application/json", nil)

	if err != nil {
		t.Fatalf("logout: %v", err)
	}

	logoutResp.Body.Close()

	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", logoutResp.StatusCode)
	}
}

func TestAuthLoginBeforeSetupReturns409(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"password": "anything"})
	resp, err := http.Post(ts.URL+"/v1/auth/login", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("login: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestAuthResumeWithUnknownTokenReturns401(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"token": "no-such-token"})
	resp, err := http.Post(ts.URL+"/v1/auth/resume", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("resume: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthWipeRefusesOtherUser(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{OSUsername: "alice"})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"osUsername": "bob"})
	resp, err := http.Post(ts.URL+"/v1/auth/wipe", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("wipe: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestAuthWipeForActiveUserReturns204(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{OSUsername: "alice"})
	ts := startHTTPServer(t, srv, false)

	body, _ := json.Marshal(map[string]string{"osUsername": "alice"})
	resp, err := http.Post(ts.URL+"/v1/auth/wipe", "application/json", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("wipe: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}
