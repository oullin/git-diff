package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigratorRunDoesNotDropTablesWhenSchemaApplyFails(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(ctx, "CREATE TABLE review_sessions (id TEXT PRIMARY KEY)"); err != nil {
		t.Fatalf("create incompatible table: %v", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE TABLE keepers (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"); err != nil {
		t.Fatalf("create keepers: %v", err)
	}

	if _, err := db.ExecContext(ctx, "INSERT INTO keepers (value) VALUES ('still here')"); err != nil {
		t.Fatalf("insert keeper: %v", err)
	}

	if err := NewMigrator(db).Run(ctx); err == nil {
		t.Fatalf("expected migration to fail instead of resetting schema")
	}

	var value string

	if err := db.QueryRowContext(ctx, "SELECT value FROM keepers WHERE id = 1").Scan(&value); err != nil {
		t.Fatalf("keeper row should remain after migration failure: %v", err)
	}

	if value != "still here" {
		t.Fatalf("keeper value = %q, want still here", value)
	}
}

func TestColumnSetReturnsQueryErrors(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	if _, err := NewMigrator(db).columnSet(ctx, "users"); err == nil {
		t.Fatalf("expected closed database error")
	}
}

func TestColumnSetQuotesTableIdentifiers(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(ctx, `CREATE TABLE "odd table "" name" (id INTEGER PRIMARY KEY, value TEXT)`); err != nil {
		t.Fatalf("create oddly named table: %v", err)
	}

	have, err := NewMigrator(db).columnSet(ctx, `odd table " name`)

	if err != nil {
		t.Fatalf("column set: %v", err)
	}

	if !have["id"] || !have["value"] {
		t.Fatalf("columns = %#v, want id and value", have)
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
