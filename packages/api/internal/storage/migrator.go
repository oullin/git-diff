package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// baselineVersion is the migration number that captures every pre-golang-migrate
// schema state. Existing user DBs without a schema_migrations table are force-marked
// at this version after legacy compatibility ALTERs are applied.

type Migrator struct {
	db *sql.DB
}

const baselineVersion = 1

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run(ctx context.Context) error {
	needsBaseline, err := m.detectLegacyDB(ctx)

	if err != nil {
		return fmt.Errorf("detect legacy db: %w", err)
	}

	if needsBaseline {
		if err := m.applyLegacyBaselineALTERs(ctx); err != nil {
			return fmt.Errorf("baseline existing db: %w", err)
		}
	}

	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if needsBaseline {
		if err := mg.Force(baselineVersion); err != nil {
			return fmt.Errorf("force baseline version: %w", err)
		}
	}

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// detectLegacyDB returns true when the DB has app tables but no
// schema_migrations — the signature of a DB created by the pre-golang-migrate
// codebase. Must run before newMigrate(), which auto-creates schema_migrations.
func (m *Migrator) detectLegacyDB(ctx context.Context) (bool, error) {
	hasMigrationsTable, err := m.tableExists(ctx, "schema_migrations")

	if err != nil {
		return false, err
	}

	if hasMigrationsTable {
		return false, nil
	}

	hasAppTable, err := m.tableExists(ctx, "users")

	if err != nil {
		return false, err
	}

	return hasAppTable, nil
}

func (m *Migrator) applyLegacyBaselineALTERs(ctx context.Context) error {
	if err := m.dropLegacyProvisioningTables(ctx); err != nil {
		return err
	}

	if err := m.addReviewSessionContextColumns(ctx); err != nil {
		return err
	}

	if err := m.addCommentRangeColumns(ctx); err != nil {
		return err
	}

	if err := m.addWalkthroughGroupsColumns(ctx); err != nil {
		return err
	}

	return m.dropWalkthroughLegacyColumns(ctx)
}

func (m *Migrator) newMigrate() (*migrate.Migrate, error) {
	src, err := iofs.New(migrationsFS, "migrations")

	if err != nil {
		return nil, fmt.Errorf("load migration source: %w", err)
	}

	drv, err := migratesqlite3.WithInstance(m.db, &migratesqlite3.Config{})

	if err != nil {
		return nil, fmt.Errorf("init sqlite3 migration driver: %w", err)
	}

	mg, err := migrate.NewWithInstance("iofs", src, "sqlite3", drv)

	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	// Callers must not invoke mg.Close(): the sqlite3 migrate driver wraps the
	// caller-owned *sql.DB and Close() would shut it down underneath us.
	return mg, nil
}

// BaselineSchemaSQL returns the raw v1 baseline migration. Useful for tools
// that need to inspect or render the expected schema without opening a DB.
func BaselineSchemaSQL() ([]byte, error) {
	return fs.ReadFile(migrationsFS, "migrations/0001_baseline.up.sql")
}

// Version reports the current schema version and whether the DB is in a dirty
// state. Returns version == 0 and ok == false when no migrations have ever run.
func (m *Migrator) Version() (version uint, dirty bool, ok bool, err error) {
	mg, err := m.newMigrate()

	if err != nil {
		return 0, false, false, err
	}

	v, d, err := mg.Version()

	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, false, nil
	}

	if err != nil {
		return 0, false, false, fmt.Errorf("read migration version: %w", err)
	}

	return v, d, true, nil
}

// Down rolls back N migration steps. Pass 0 to roll back everything.
func (m *Migrator) Down(steps int) error {
	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if steps <= 0 {
		if err := mg.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate down: %w", err)
		}

		return nil
	}

	if err := mg.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down %d: %w", steps, err)
	}

	return nil
}

// Force marks the DB at version v with dirty=false. Used to recover from a
// failed migration after the operator has manually inspected the state.
func (m *Migrator) Force(version int) error {
	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if err := mg.Force(version); err != nil {
		return fmt.Errorf("force version %d: %w", version, err)
	}

	return nil
}

func (m *Migrator) tableExists(ctx context.Context, name string) (bool, error) {
	var got string

	err := m.db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
		name,
	).Scan(&got)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("query sqlite_master for %s: %w", name, err)
	}

	return true, nil
}

func (m *Migrator) addReviewSessionContextColumns(ctx context.Context) error {
	have, err := m.columnSet(ctx, "review_sessions")

	if err != nil {
		return err
	}

	if len(have) == 0 {
		return nil
	}

	if !have["context_kind"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier("review_sessions")+" ADD COLUMN context_kind TEXT NOT NULL DEFAULT 'working'"); err != nil {
			return fmt.Errorf("add context_kind: %w", err)
		}
	}

	if !have["context_sha"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier("review_sessions")+" ADD COLUMN context_sha TEXT"); err != nil {
			return fmt.Errorf("add context_sha: %w", err)
		}
	}

	return nil
}

func (m *Migrator) addCommentRangeColumns(ctx context.Context) error {
	for _, table := range []string{"review_comments", "pending_comments"} {
		have, err := m.columnSet(ctx, table)

		if err != nil {
			return err
		}

		if len(have) == 0 {
			continue
		}

		if !have["start_line_number"] {
			if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier(table)+" ADD COLUMN start_line_number INTEGER"); err != nil {
				return fmt.Errorf("add %s.start_line_number: %w", table, err)
			}
		}

		if !have["start_side"] {
			if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier(table)+" ADD COLUMN start_side TEXT"); err != nil {
				return fmt.Errorf("add %s.start_side: %w", table, err)
			}
		}
	}

	return nil
}

func (m *Migrator) addWalkthroughGroupsColumns(ctx context.Context) error {
	have, err := m.columnSet(ctx, "walkthroughs")

	if err != nil {
		return err
	}

	if len(have) == 0 {
		return nil
	}

	if !have["groups_json"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier("walkthroughs")+" ADD COLUMN groups_json TEXT NOT NULL DEFAULT '[]'"); err != nil {
			return fmt.Errorf("add walkthroughs.groups_json: %w", err)
		}
	}

	if !have["provider_id"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier("walkthroughs")+" ADD COLUMN provider_id TEXT NOT NULL DEFAULT ''"); err != nil {
			return fmt.Errorf("add walkthroughs.provider_id: %w", err)
		}
	}

	return nil
}

// dropWalkthroughLegacyColumns removes the pre-groups `order_json` / `notes_json`
// columns from legacy SQLite DBs. The current writer never supplies values for
// them, so leaving the NOT NULL columns in place would break inserts.
func (m *Migrator) dropWalkthroughLegacyColumns(ctx context.Context) error {
	have, err := m.columnSet(ctx, "walkthroughs")

	if err != nil {
		return err
	}

	if len(have) == 0 {
		return nil
	}

	for _, column := range []string{"order_json", "notes_json"} {
		if !have[column] {
			continue
		}

		if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+sqliteQuoteIdentifier("walkthroughs")+" DROP COLUMN "+sqliteQuoteIdentifier(column)); err != nil {
			return fmt.Errorf("drop walkthroughs.%s: %w", column, err)
		}
	}

	return nil
}

func (m *Migrator) dropLegacyProvisioningTables(ctx context.Context) error {
	for _, table := range []string{"workflow_events", "workflow_runs"} {
		if _, err := m.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+sqliteQuoteIdentifier(table)); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}

	return nil
}

func (m *Migrator) columnSet(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := m.db.QueryContext(ctx, "PRAGMA table_info("+sqliteQuoteIdentifier(table)+")")

	if err != nil {
		return nil, fmt.Errorf("query %s columns: %w", table, err)
	}

	defer rows.Close()

	have := map[string]bool{}

	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("scan %s column row: %w", table, err)
		}

		have[name] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s column rows: %w", table, err)
	}

	return have, nil
}

func sqliteQuoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
