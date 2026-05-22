package userconfig

import (
	"context"
	"fmt"
	"path/filepath"
)

// Service bundles the loader, reader, watcher, and broker into one
// dependency the HTTP layer can inject. The struct is just composition —
// every concern still lives in its own file.
//
// Construct via NewService at startup. Call Run(ctx) in a goroutine to
// drive the watcher; cancel ctx to stop.
type Service struct {
	Path    string
	Reader  *AtomicReader
	Broker  *Broker
	watcher *Watcher
}

// NewService scaffolds the YAML file if missing, loads it, and returns a
// Service ready to be Run. Returns the load error so callers can fail
// loudly at boot.
func NewService(home string) (*Service, error) {
	path := DefaultPath(home)

	if err := WriteDefaults(path); err != nil {
		return nil, fmt.Errorf("scaffold user config: %w", err)
	}

	cfg, err := Loader{}.Load(path)

	if err != nil {
		return nil, fmt.Errorf("load user config: %w", err)
	}

	return &Service{
		Path:    path,
		Reader:  NewAtomicReader(cfg),
		Broker:  NewBroker(),
		watcher: NewWatcher(),
	}, nil
}

// Run drives the watcher loop. On every reload, the AtomicReader is
// updated and the broker fans out to subscribers. Errors are silently
// dropped here — production callers should pass a logger via onError;
// kept off the surface to avoid coupling to a logging library.
func (s *Service) Run(ctx context.Context) error {
	return s.watcher.Watch(ctx, s.Path,
		func(cfg Config) {
			s.Reader.Set(cfg)
			s.Broker.Push(cfg)
		},
		nil,
	)
}

// DefaultPath returns the canonical config location for a given home
// directory: ~/.git-diff/config.yaml.
func DefaultPath(home string) string {
	return filepath.Join(home, ".git-diff", "config.yaml")
}
