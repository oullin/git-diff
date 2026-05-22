package userconfig

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcherEmitsInitialLoadThenDebouncesEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := os.WriteFile(path, []byte("theme: dark\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var (
		updates  atomic.Int32
		lastDark atomic.Bool
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	w := &Watcher{loader: Loader{}, debounce: 50 * time.Millisecond}

	done := make(chan struct{})

	go func() {
		_ = w.Watch(ctx, path, func(cfg Config) {
			updates.Add(1)

			if cfg.Theme == "dark" {
				lastDark.Store(true)
			} else {
				lastDark.Store(false)
			}
		}, nil)

		close(done)
	}()

	waitFor(t, "initial load", func() bool { return updates.Load() == 1 })

	// Bash three rapid writes — the debounce should coalesce them.
	for i := 0; i < 3; i++ {
		if err := os.WriteFile(path, []byte("theme: light\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	waitFor(t, "debounced reload", func() bool { return updates.Load() >= 2 && !lastDark.Load() })

	if got := updates.Load(); got > 4 {
		t.Fatalf("debounce failed: %d updates for 3 writes + 1 initial", got)
	}

	cancel()
	<-done
}

func TestWatcherSurfacesParseErrorsWithoutKillingTheLoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := os.WriteFile(path, []byte("theme: dark\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var (
		updates atomic.Int32
		errs    atomic.Int32
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	w := &Watcher{loader: Loader{}, debounce: 30 * time.Millisecond}

	go func() {
		_ = w.Watch(ctx, path,
			func(Config) { updates.Add(1) },
			func(error) { errs.Add(1) },
		)
	}()

	waitFor(t, "initial load", func() bool { return updates.Load() >= 1 })

	if err := os.WriteFile(path, []byte("theme: dark\n  garbage: : :\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "parse error reported", func() bool { return errs.Load() >= 1 })

	// Recovering should resume notifications — loop is still alive.
	if err := os.WriteFile(path, []byte("theme: system\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "recovered update", func() bool { return updates.Load() >= 2 })
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("timeout waiting for %s", what)
}
