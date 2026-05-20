package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gocanto/git-diff/internal/command"
)

type stubRunner struct {
	outputs map[string][]byte
	errors  map[string]error
	calls   *[]string
}

func (r stubRunner) Run(name string, args ...string) ([]byte, error) {
	key := command.ShellQuote(append([]string{name}, args...))

	if r.calls != nil {
		*r.calls = append(*r.calls, key)
	}

	return r.outputs[key], r.errors[key]
}

func TestNoArgsShowsUsage(t *testing.T) {
	var stdout bytes.Buffer
	a := newApp("/Users/gus", "/repo", strings.NewReader(""), io.Discard, io.Discard, stubRunner{})
	a.stdout = &stdout

	if got := a.run(nil); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}

	if !strings.Contains(stdout.String(), "api serve-http --socket") {
		t.Fatalf("usage = %s", stdout.String())
	}
}

func TestUnknownCommandsAreRejected(t *testing.T) {
	for _, cmd := range []string{"doctor", "list-workflows", "run-workflow", "brewfile"} {
		t.Run(cmd, func(t *testing.T) {
			var stderr bytes.Buffer
			a := newApp("/Users/gus", "/repo", strings.NewReader(""), io.Discard, &stderr, stubRunner{})

			if got := a.run([]string{cmd}); got != 2 {
				t.Fatalf("exit = %d, want 2", got)
			}

			if !strings.Contains(stderr.String(), `unknown command "`+cmd+`"`) {
				t.Fatalf("stderr = %s", stderr.String())
			}
		})
	}
}

func TestHelpOnlyShowsServeHTTPUsage(t *testing.T) {
	var stdout bytes.Buffer
	a := newApp("/Users/gus", "/repo", strings.NewReader(""), &stdout, io.Discard, stubRunner{})

	if got := a.run([]string{"help"}); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}

	output := stdout.String()

	for _, want := range []string{"api", "serve-http", "Unix HTTP socket"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output = %s, want %q", output, want)
		}
	}

	for _, gone := range []string{"list-workflows", "run-workflow", "bootstrap", "adopt", "restore", "secrets", "doctor", "brewfile", "macos"} {
		if strings.Contains(output, gone) {
			t.Fatalf("help output = %s, did not expect removed command %q", output, gone)
		}
	}
}

func TestFindRepoRootWalksUp(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(dir, "cmd", "api")

	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}

	if got := findRepoRoot(nested); got != dir {
		t.Fatalf("findRepoRoot(%q) = %q, want %q", nested, got, dir)
	}
}

