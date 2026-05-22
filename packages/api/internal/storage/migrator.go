package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
)

//go:embed schema.sql
var schemaFS embed.FS

// Migrator runs explicit compatibility migrations before applying the
// embedded schema.
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

	if err := m.addWalkthroughGroupsColumns(ctx); err != nil {
		return fmt.Errorf("add walkthrough groups columns: %w", err)
	}

	if _, err := m.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("initialize sqlite schema: %w", err)
	}

	return nil
}

// CREATE TABLE IF NOT EXISTS is a no-op on existing tables, so legacy
// databases need explicit ALTERs to pick up the new columns.
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

// Multi-line and cross-side comment ranges live in start_line_number /
// start_side; single-line comments leave both NULL.
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

// groups_json defaults to "[]" so cached rows without it still decode cleanly.
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

func (m *Migrator) dropLegacyProvisioningTables(ctx context.Context) error {
	for _, table := range []string{"workflow_events", "workflow_runs"} {
		if _, err := m.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+sqliteQuoteIdentifier(table)); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}

	return nil
}

// columnSet returns the columns present on table, or an empty set if the table
// does not exist.
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
