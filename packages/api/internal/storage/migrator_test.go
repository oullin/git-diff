package storage

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := url.URL{Scheme: "file", Path: filepath.Join(t.TempDir(), "test.sqlite3")}
	q := dsn.Query()
	q.Add("_foreign_keys", "on")
	dsn.RawQuery = q.Encode()

	db, err := sql.Open("sqlite3", dsn.String())

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func TestMigratorRunOnFreshDB(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("migrate run: %v", err)
	}

	for _, table := range []string{"users", "review_sessions", "walkthroughs", "pending_comments", "branches", "schema_migrations"} {
		exists, err := NewMigrator(db).tableExists(ctx, table)

		if err != nil {
			t.Fatalf("tableExists %s: %v", table, err)
		}

		if !exists {
			t.Fatalf("expected table %s to exist after fresh migrate", table)
		}
	}

	v, dirty, ok, err := NewMigrator(db).Version()

	if err != nil {
		t.Fatalf("version: %v", err)
	}

	if !ok || v != 1 || dirty {
		t.Fatalf("version = %d dirty=%t ok=%t, want version=1 clean ok=true", v, dirty, ok)
	}
}

func TestMigratorRunIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("first run: %v", err)
	}

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("second run: %v", err)
	}
}

func TestMigratorVersionEmptyDB(t *testing.T) {
	db := openTestDB(t)
	_, _, ok, err := NewMigrator(db).Version()

	if err != nil {
		t.Fatalf("version: %v", err)
	}

	if ok {
		t.Fatalf("expected ok=false on empty DB")
	}
}

func TestMigratorDownRollsBackBaseline(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("migrate run: %v", err)
	}

	if err := NewMigrator(db).Down(0); err != nil {
		t.Fatalf("migrate down: %v", err)
	}

	exists, err := NewMigrator(db).tableExists(ctx, "users")

	if err != nil {
		t.Fatalf("tableExists users: %v", err)
	}

	if exists {
		t.Fatalf("users should have been dropped after full down migrate")
	}
}

func TestOpenEnablesForeignKeysOnEachConnection(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })
	store.db.SetMaxOpenConns(2)

	first, err := store.db.Conn(ctx)

	if err != nil {
		t.Fatalf("first conn: %v", err)
	}

	defer first.Close()

	second, err := store.db.Conn(ctx)

	if err != nil {
		t.Fatalf("second conn: %v", err)
	}

	defer second.Close()

	for name, conn := range map[string]*sql.Conn{"first": first, "second": second} {
		var enabled int

		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil {
			t.Fatalf("%s conn pragma: %v", name, err)
		}

		if enabled != 1 {
			t.Fatalf("%s conn foreign_keys = %d, want 1", name, enabled)
		}
	}

	_, err = second.ExecContext(ctx, "INSERT INTO user_sessions (token, user_id, created_at, expires_at, last_used_at) VALUES ('bad', 999, '', '', '')")

	if err == nil {
		t.Fatalf("expected foreign key violation on second connection")
	}
}
