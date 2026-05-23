package httpx

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestUserConfigGetReturnsDefaults(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/userconfig")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var cfg map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// usercfg.Defaults() ships with these values.
	if cfg["theme"] != "system" {
		t.Fatalf("expected theme=system, got %#v", cfg["theme"])
	}

	// The handler must surface the config-file path so the renderer can
	// pass it to shell.openPath without re-deriving it.
	path, ok := cfg["path"].(string)

	if !ok || path == "" {
		t.Fatalf("expected non-empty path, got %#v", cfg["path"])
	}

	if !strings.HasSuffix(path, "/.git-diff/config.yaml") {
		t.Fatalf("expected path ending in /.git-diff/config.yaml, got %q", path)
	}

	walkthrough, ok := cfg["walkthrough"].(map[string]any)

	if !ok {
		t.Fatalf("expected walkthrough block, got %#v", cfg["walkthrough"])
	}

	if walkthrough["provider"] != "anthropic" {
		t.Fatalf("expected provider=anthropic, got %#v", walkthrough["provider"])
	}
}

func TestUserConfigGetReturnsServiceUnavailableWhenReaderNil(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	srv.UserConfig = nil
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/userconfig")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}

func TestUserConfigStreamEmitsInitialEvent(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	ts := startHTTPServer(t, srv, false)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/v1/userconfig/stream", nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("expected text/event-stream Content-Type, got %q", ct)
	}

	// Read the first SSE frame on a deadline; the handler emits the initial
	// snapshot synchronously, so this should not hang.
	scanner := bufio.NewScanner(resp.Body)
	done := make(chan struct{})
	gotEvent := false

	var dataLine string

	go func() {
		defer close(done)

		sawEvent := false

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "event: config") {
				gotEvent = true
				sawEvent = true

				continue
			}

			if sawEvent && strings.HasPrefix(line, "data: ") {
				dataLine = strings.TrimPrefix(line, "data: ")

				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("did not receive initial SSE frame within 2s")
	}

	if !gotEvent {
		t.Fatalf("expected an `event: config` line on the stream")
	}

	if dataLine == "" {
		t.Fatalf("expected a `data:` line after `event: config`")
	}

	var payload map[string]any

	if err := json.Unmarshal([]byte(dataLine), &payload); err != nil {
		t.Fatalf("decode SSE data: %v\n%s", err, dataLine)
	}

	if path, _ := payload["path"].(string); path == "" {
		t.Fatalf("expected SSE payload to include path, got %#v", payload["path"])
	}
}

func TestUserConfigStreamReturnsServiceUnavailableWhenBrokerNil(t *testing.T) {
	srv, _ := testServer(t, testServerOptions{})
	srv.UserConfigEvents = nil
	ts := startHTTPServer(t, srv, false)

	resp, err := http.Get(ts.URL + "/v1/userconfig/stream")

	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}
