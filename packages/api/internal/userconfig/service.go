package userconfig

import (
	"context"
	"fmt"
	"path/filepath"
)

// Service bundles loader, reader, watcher, and broker. Construct via
// NewService at startup; call Run(ctx) in a goroutine and cancel to stop.
type Service struct {
	Path    string
	Reader  *AtomicReader
	Broker  *Broker
	watcher *Watcher
}

// NewService scaffolds the YAML file if missing and returns a load error
// so callers can fail loudly at boot.
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

// Run drives the watcher loop, refreshing the reader and broker on every
// reload. Watch errors are silently dropped; pass a logger via the
// watcher directly if production logging is needed.
func (s *Service) Run(ctx context.Context) error {
	return s.watcher.Watch(ctx, s.Path,
		func(cfg Config) {
			s.Reader.Set(cfg)
			s.Broker.Push(cfg)
		},
		nil,
	)
}

// DefaultPath is ~/.git-diff/config.yaml relative to home.
func DefaultPath(home string) string {
	return filepath.Join(home, ".git-diff", "config.yaml")
}
