package userconfig

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBrokerFansOutToAllSubscribers(t *testing.T) {
	b := NewBroker()

	chA, unsubA := b.Subscribe()

	defer unsubA()

	chB, unsubB := b.Subscribe()

	defer unsubB()

	cfg := Defaults()
	cfg.Theme = "dark"

	go b.Push(cfg)

	got := <-chA

	if got.Theme != "dark" {
		t.Fatalf("subscriber A got %q", got.Theme)
	}

	got = <-chB

	if got.Theme != "dark" {
		t.Fatalf("subscriber B got %q", got.Theme)
	}
}

func TestBrokerUnsubscribeStopsDelivery(t *testing.T) {
	b := NewBroker()
	ch, unsub := b.Subscribe()
	unsub()

	// Channel must be closed; reading must not block.
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed channel to read the zero value")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("read on unsubscribed channel blocked — channel was not closed")
	}

	// Push after unsubscribe must not panic.
	b.Push(Defaults())
}

func TestBrokerDropsForSlowSubscribers(t *testing.T) {
	b := NewBroker()
	_, unsub := b.Subscribe()

	defer unsub()

	// Fill the buffer (size 1), then push more — extras must drop, not block.
	done := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			b.Push(Defaults())
		}

		close(done)
	}()

	select {
	case <-done:
		// good — Push didn't block.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Push blocked despite slow consumer")
	}
}

func TestBrokerDoubleUnsubscribeIsSafe(t *testing.T) {
	b := NewBroker()
	_, unsub := b.Subscribe()

	var panics atomic.Bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				panics.Store(true)
			}
		}()

		unsub()
		unsub()
	}()

	if panics.Load() {
		t.Fatal("double unsubscribe should be safe")
	}
}
