package userconfig

import "sync/atomic"

// Reader is the read-only surface handlers depend on.
type Reader interface {
	Get() Config
}

// AtomicReader carries the config through a lock-free pointer swap so the
// watcher goroutine can publish without blocking readers.
type AtomicReader struct {
	value atomic.Pointer[Config]
}

func NewAtomicReader(cfg Config) *AtomicReader {
	r := &AtomicReader{}
	r.Set(cfg)

	return r
}

func (r *AtomicReader) Get() Config {
	if v := r.value.Load(); v != nil {
		return *v
	}

	return Defaults()
}

func (r *AtomicReader) Set(cfg Config) {
	r.value.Store(&cfg)
}
