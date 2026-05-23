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

	if !ok || v != baselineVersion || dirty {
		t.Fatalf("version = %d dirty=%t ok=%t, want version=%d clean ok=true", v, dirty, ok, baselineVersion)
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

// Simulates a legacy DB created by an older binary: tables exist but the
// additive columns from the four compat ALTERs are missing, and the legacy
// workflow_events / workflow_runs tables are still present. The migrator must
// converge it to v1 without re-running the baseline migration.
func TestMigratorBaselinesLegacyDB(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	preALTERSchema := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			os_username TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			last_login_at TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE review_sessions (
			id TEXT PRIMARY KEY,
			repo_root TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			branch TEXT NOT NULL,
			head_sha TEXT NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			files_changed INTEGER NOT NULL DEFAULT 0,
			additions INTEGER NOT NULL DEFAULT 0,
			deletions INTEGER NOT NULL DEFAULT 0,
			started_at TEXT NOT NULL,
			completed_at TEXT
		)`,
		`CREATE TABLE review_comments (
			id TEXT PRIMARY KEY,
			review_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			diff_section TEXT NOT NULL,
			side TEXT NOT NULL,
			line_number INTEGER NOT NULL,
			author_label TEXT NOT NULL,
			body_html TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
		)`,
		`CREATE TABLE pending_comments (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			repo_root TEXT NOT NULL,
			context_kind TEXT NOT NULL DEFAULT 'working',
			context_sha TEXT NOT NULL DEFAULT '',
			file_path TEXT NOT NULL,
			diff_section TEXT NOT NULL,
			side TEXT NOT NULL,
			line_number INTEGER NOT NULL,
			author_label TEXT NOT NULL,
			body_html TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE walkthroughs (
			repo_root TEXT NOT NULL,
			context_kind TEXT NOT NULL,
			context_sha TEXT NOT NULL DEFAULT '',
			fingerprint TEXT NOT NULL,
			model_id TEXT NOT NULL,
			order_json TEXT NOT NULL,
			notes_json TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			generated_at TEXT NOT NULL,
			PRIMARY KEY (repo_root, context_kind, context_sha)
		)`,
		`CREATE TABLE workflow_runs (id INTEGER PRIMARY KEY)`,
		`CREATE TABLE workflow_events (id INTEGER PRIMARY KEY)`,
	}

	for _, stmt := range preALTERSchema {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed legacy schema: %v", err)
		}
	}

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("migrate run: %v", err)
	}

	mig := NewMigrator(db)

	for _, table := range []string{"workflow_events", "workflow_runs"} {
		exists, err := mig.tableExists(ctx, table)

		if err != nil {
			t.Fatalf("tableExists %s: %v", table, err)
		}

		if exists {
			t.Fatalf("legacy table %s should have been dropped", table)
		}
	}

	expectedColumns := map[string][]string{
		"review_sessions":  {"context_kind", "context_sha"},
		"review_comments":  {"start_line_number", "start_side"},
		"pending_comments": {"start_line_number", "start_side"},
		"walkthroughs":     {"groups_json", "provider_id"},
	}

	for table, cols := range expectedColumns {
		have, err := mig.columnSet(ctx, table)

		if err != nil {
			t.Fatalf("columnSet %s: %v", table, err)
		}

		for _, col := range cols {
			if !have[col] {
				t.Fatalf("%s.%s missing after baseline", table, col)
			}
		}
	}

	v, dirty, ok, err := mig.Version()

	if err != nil {
		t.Fatalf("version: %v", err)
	}

	if !ok || v != baselineVersion || dirty {
		t.Fatalf("version = %d dirty=%t ok=%t, want version=%d clean ok=true", v, dirty, ok, baselineVersion)
	}
}

func TestMigratorBaselinesFullyMigratedLegacyDB(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	baseline, err := BaselineSchemaSQL()

	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(baseline)); err != nil {
		t.Fatalf("seed full baseline schema: %v", err)
	}

	if err := NewMigrator(db).Run(ctx); err != nil {
		t.Fatalf("migrate run: %v", err)
	}

	v, _, ok, err := NewMigrator(db).Version()

	if err != nil {
		t.Fatalf("version: %v", err)
	}

	if !ok || v != baselineVersion {
		t.Fatalf("version = %d ok=%t, want %d ok=true", v, ok, baselineVersion)
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

func TestColumnSetReturnsQueryErrors(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	if _, err := NewMigrator(db).columnSet(ctx, "users"); err == nil {
		t.Fatalf("expected closed database error")
	}
}

func TestColumnSetQuotesTableIdentifiers(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

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
