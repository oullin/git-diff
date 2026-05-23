// Package db owns the process-wide gorm connection and clock used by the
// storage layer. Repos read both via Conn() and Now() instead of carrying
// per-instance fields.
package db

import (
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
)

var (
	mu   sync.RWMutex
	conn *gorm.DB
	now  = time.Now
)

// Init installs the process-wide gorm connection. Called once from storage.Open.
func Init(c *gorm.DB) {
	mu.Lock()

	defer mu.Unlock()

	conn = c
}

// Conn returns the global connection. Panics if Init has not been called —
// failing loudly here is preferable to a nil-deref deep inside a query.
func Conn() *gorm.DB {
	mu.RLock()

	defer mu.RUnlock()

	if conn == nil {
		panic("db: Conn called before Init")
	}

	return conn
}

// Now returns the current time through the swappable clock.
func Now() time.Time {
	mu.RLock()

	defer mu.RUnlock()

	return now()
}

// SetNow swaps the clock. Used by tests via Store.SetNow.
func SetNow(fn func() time.Time) {
	mu.Lock()

	defer mu.Unlock()

	now = fn
}

// SetForTest swaps conn (and resets the clock to time.Now) for the duration of
// the test, restoring the previous values on t.Cleanup. Tests must not use
// t.Parallel() — the globals are shared process-wide.
func SetForTest(t testing.TB, c *gorm.DB) {
	t.Helper()

	mu.Lock()
	prevConn, prevNow := conn, now
	conn, now = c, time.Now
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()

		defer mu.Unlock()

		conn, now = prevConn, prevNow
	})
}
