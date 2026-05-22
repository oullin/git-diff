package userconfig

import (
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher emits debounced notifications when the YAML file changes on
// disk. Single responsibility: notify. Does NOT decode, validate, or
// store state — the callback supplies the parsing.
//
// Debounced at 200 ms to coalesce editor saves (vim's write-and-rename,
// VS Code's atomic save, etc.) that would otherwise fire the callback
// 2-3 times per save.
type Watcher struct {
	loader   Loader
	debounce time.Duration
}

// NewWatcher returns a watcher with the default 200 ms debounce.
func NewWatcher() *Watcher {
	return &Watcher{loader: Loader{}, debounce: 200 * time.Millisecond}
}

// Watch blocks until ctx is cancelled, invoking onChange(cfg) after each
// debounced file edit. Returns immediately with the first load error
// (e.g. the watch target was missing); transient parse failures after
// startup are surfaced through onError so the watcher loop keeps running.
func (w *Watcher) Watch(
	ctx context.Context,
	path string,
	onChange func(Config),
	onError func(error),
) error {
	notifier, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	defer notifier.Close()

	if err := notifier.Add(path); err != nil {
		return err
	}

	// Send the initial load synchronously so subscribers don't have to
	// race the first edit to learn what's on disk.
	cfg, err := w.loader.Load(path)

	if err != nil {
		return err
	}

	onChange(cfg)

	var (
		mu       sync.Mutex
		pending  *time.Timer
		schedule = func() {
			mu.Lock()

			defer mu.Unlock()

			if pending != nil {
				pending.Stop()
			}

			pending = time.AfterFunc(w.debounce, func() {
				cfg, err := w.loader.Load(path)

				if err != nil {
					if onError != nil {
						onError(err)
					}

					return
				}

				onChange(cfg)
			})
		}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-notifier.Events:
			if !ok {
				return nil
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				schedule()
			}
		case watchErr, ok := <-notifier.Errors:
			if !ok {
				return nil
			}

			if onError != nil {
				onError(watchErr)
			}
		}
	}
}
