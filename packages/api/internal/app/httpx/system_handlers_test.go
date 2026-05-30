package httpx

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSystemStatsEndpointReturnsJSONShape(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/system/stats")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	// The numbers are platform-dependent (macOS uses top/vm_stat). We only
	// assert that the JSON shape decodes successfully, which is what the API
	// contract promises clients.
	var stats SystemStats

	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
