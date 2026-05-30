package usercfg

import (
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher coalesces editor saves (vim's write-and-rename, VS Code's
// atomic save) that would otherwise fire the callback multiple times per
// save. 200 ms debounce.
type Watcher struct {
	loader   Loader
	debounce time.Duration
}

func NewWatcher() *Watcher {
	return &Watcher{loader: Loader{}, debounce: 200 * time.Millisecond}
}

// Watch blocks until ctx is cancelled. Returns the first load error
// immediately; later parse failures go through onError so the loop keeps
// running.
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

	// Initial load is synchronous so subscribers see current state
	// without waiting for the first edit.
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
