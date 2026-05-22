package userconfig

import "sync/atomic"

// Reader is the narrow read-only surface handlers depend on. Decouples
// callers from the loader/watcher concretes (Dependency Inversion) — most
// consumers only need "what's the current config right now?".
type Reader interface {
	Get() Config
}

// AtomicReader is a Reader backed by an atomic pointer so updates from
// the watcher goroutine are visible to handler goroutines without locks.
//
// Construct one, install it as the system's Reader, then call Set from
// the watcher whenever the file changes.
type AtomicReader struct {
	value atomic.Pointer[Config]
}

// NewAtomicReader returns a reader pre-loaded with cfg. Callers should
// pass a complete Config (typically Loader.Load output merged with
// Defaults) — Get returns whatever was last Set.
func NewAtomicReader(cfg Config) *AtomicReader {
	r := &AtomicReader{}
	r.Set(cfg)

	return r
}

// Get returns a value-typed copy of the current config. Safe to call from
// any goroutine; the underlying pointer swap is lock-free.
func (r *AtomicReader) Get() Config {
	if v := r.value.Load(); v != nil {
		return *v
	}

	return Defaults()
}

// Set replaces the current config. The next Get will return the new value.
func (r *AtomicReader) Set(cfg Config) {
	r.value.Store(&cfg)
}
