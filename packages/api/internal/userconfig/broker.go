package userconfig

import "sync"

// Broker fans out config updates from the watcher to many subscribers
// (typically SSE-streaming HTTP handlers). One-way: publishers Push,
// subscribers Subscribe and read until they Unsubscribe.
//
// Decoupled from the Watcher and Reader so any source can publish — keeps
// the SSE handler oblivious to the file-watching mechanics.
type Broker struct {
	mu          sync.Mutex
	subscribers map[chan Config]struct{}
}

// NewBroker returns an empty broker. Safe to use from any goroutine.
func NewBroker() *Broker {
	return &Broker{subscribers: make(map[chan Config]struct{})}
}

// Subscribe registers a channel that receives subsequent Push values.
// The returned unsubscribe func must be called when the subscriber goes
// away — typically `defer unsubscribe()` in the handler.
//
// The channel is buffered (size 1) so a slow consumer drops at most one
// stale update without blocking the publisher.
func (b *Broker) Subscribe() (<-chan Config, func()) {
	ch := make(chan Config, 1)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()

		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}

		b.mu.Unlock()
	}

	return ch, unsubscribe
}

// Push delivers cfg to every active subscriber. Non-blocking: if a
// subscriber's buffer is full (slow reader), the new value is dropped
// for that subscriber only.
func (b *Broker) Push(cfg Config) {
	b.mu.Lock()

	defer b.mu.Unlock()

	for ch := range b.subscribers {
		select {
		case ch <- cfg:
		default:
			// Subscriber is behind by at least one update; drop and move on.
		}
	}
}
