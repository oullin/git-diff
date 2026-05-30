package usercfg

import (
	"sync"
	"testing"
)

func TestAtomicReaderRoundTrip(t *testing.T) {
	r := NewAtomicReader(Defaults())

	if r.Get().Theme != "system" {
		t.Fatal("initial value wrong")
	}

	updated := Defaults()
	updated.Theme = "dark"
	r.Set(updated)

	if r.Get().Theme != "dark" {
		t.Fatal("update not visible")
	}
}

func TestAtomicReaderConcurrentReadsAndWrites(t *testing.T) {
	// 100 readers + 10 writers shouldn't race or panic. The atomic pointer
	// guarantees a coherent Config view at every Get; we just verify the
	// happy path doesn't deadlock or trigger the race detector.
	r := NewAtomicReader(Defaults())

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_ = r.Get()
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			cfg := Defaults()
			cfg.Theme = "dark"
			r.Set(cfg)
		}(i)
	}

	wg.Wait()
}
