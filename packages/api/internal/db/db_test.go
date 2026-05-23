package db

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnPanicsBeforeInit(t *testing.T) {
	// Snapshot and clear the package globals so the test sees a fresh state,
	// then restore them on cleanup so the rest of the suite keeps working.
	mu.Lock()
	prev := conn
	conn = nil
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()

		defer mu.Unlock()

		conn = prev
	})

	defer func() {
		if recover() == nil {
			t.Fatalf("expected Conn() to panic before Init")
		}
	}()

	_ = Conn()
}

func TestInitAndConnReturnsSameInstance(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "init.sqlite3")), &gorm.Config{})

	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	SetForTest(t, gdb)

	if Conn() != gdb {
		t.Fatalf("Conn() returned a different connection than the one installed via SetForTest")
	}
}

func TestNowSetNowSwappable(t *testing.T) {
	fixed := time.Date(2030, 1, 2, 3, 4, 5, 6, time.UTC)

	prev := now
	SetNow(func() time.Time { return fixed })

	t.Cleanup(func() { SetNow(prev) })

	if got := Now(); !got.Equal(fixed) {
		t.Fatalf("Now() = %v, want %v", got, fixed)
	}
}
