package userconfig

import "sync"

// Broker fans config updates out to SSE subscribers.
type Broker struct {
	mu          sync.Mutex
	subscribers map[chan Config]struct{}
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[chan Config]struct{})}
}

// Subscribe returns a buffered channel (size 1) so a slow consumer drops
// at most one stale update without blocking the publisher. The caller
// must invoke the returned unsubscribe func when done.
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

// Push is non-blocking; a subscriber whose buffer is full loses the update.
func (b *Broker) Push(cfg Config) {
	b.mu.Lock()

	defer b.mu.Unlock()

	for ch := range b.subscribers {
		select {
		case ch <- cfg:
		default:
		}
	}
}
