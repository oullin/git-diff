package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed schema.sql
var schemaFS embed.FS

// Migrator brings a SQLite database up to the current schema. It runs the
// embedded `schema.sql`, applies one-shot column migrations for legacy
// databases, and falls back to a destructive reset if the embedded schema
// fails to apply (e.g. when an old table shape blocks new constraints).
type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run(ctx context.Context) error {
	schema, err := schemaFS.ReadFile("schema.sql")

	if err != nil {
		return fmt.Errorf("read embedded sqlite schema: %w", err)
	}

	if err := m.dropLegacyProvisioningTables(ctx); err != nil {
		return fmt.Errorf("drop legacy provisioning tables: %w", err)
	}

	if err := m.addReviewSessionContextColumns(ctx); err != nil {
		return fmt.Errorf("add review_sessions context columns: %w", err)
	}

	if err := m.addCommentRangeColumns(ctx); err != nil {
		return fmt.Errorf("add comment range columns: %w", err)
	}

	if _, err := m.db.ExecContext(ctx, string(schema)); err == nil {
		return nil
	}

	if err := m.resetSchema(ctx); err != nil {
		return fmt.Errorf("reset stale schema: %w", err)
	}

	if _, err := m.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("initialize sqlite schema after reset: %w", err)
	}

	return nil
}

// addReviewSessionContextColumns is a one-shot migration that brings legacy
// databases up to the current schema by attaching the context_kind and
// context_sha columns to review_sessions. CREATE TABLE IF NOT EXISTS in
// schema.sql is a no-op for existing tables, so explicit ALTERs are required.
func (m *Migrator) addReviewSessionContextColumns(ctx context.Context) error {
	have, err := m.columnSet(ctx, "review_sessions")

	if err != nil {
		return err
	}

	if len(have) == 0 {
		return nil
	}

	if !have["context_kind"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE review_sessions ADD COLUMN context_kind TEXT NOT NULL DEFAULT 'working'"); err != nil {
			return fmt.Errorf("add context_kind: %w", err)
		}
	}

	if !have["context_sha"] {
		if _, err := m.db.ExecContext(ctx, "ALTER TABLE review_sessions ADD COLUMN context_sha TEXT"); err != nil {
			return fmt.Errorf("add context_sha: %w", err)
		}
	}

	return nil
}

// addCommentRangeColumns brings legacy databases up to the current schema by
// attaching start_line_number / start_side columns to review_comments and
// pending_comments. Multi-line and cross-side comment ranges are stored
// through these fields; single-line comments leave both NULL.
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
			if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN start_line_number INTEGER"); err != nil {
				return fmt.Errorf("add %s.start_line_number: %w", table, err)
			}
		}

		if !have["start_side"] {
			if _, err := m.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN start_side TEXT"); err != nil {
				return fmt.Errorf("add %s.start_side: %w", table, err)
			}
		}
	}

	return nil
}

// dropLegacyProvisioningTables removes tables that used to back the macOS
// provisioning engine. They are no longer part of the schema; this lets
// existing user databases shed them on first launch after the upgrade.
func (m *Migrator) dropLegacyProvisioningTables(ctx context.Context) error {
	for _, table := range []string{"workflow_events", "workflow_runs"} {
		if _, err := m.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}

	return nil
}

func (m *Migrator) resetSchema(ctx context.Context) error {
	if _, err := m.db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	rows, err := m.db.QueryContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
	`)

	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}

	tables := []string{}

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			rows.Close()

			return fmt.Errorf("scan table name: %w", err)
		}

		tables = append(tables, name)
	}

	rows.Close()

	for _, table := range tables {
		if _, err := m.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("drop table %q: %w", table, err)
		}
	}

	if _, err := m.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}

// columnSet returns the set of column names present on `table`, or an empty
// set if the table does not exist. Any SQL error is treated as "table not
// present yet" by the caller (the schema.sql apply will create it).
func (m *Migrator) columnSet(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := m.db.QueryContext(ctx, "PRAGMA table_info("+table+")")

	if err != nil {
		return nil, nil
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
