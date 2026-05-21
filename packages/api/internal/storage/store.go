package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
	_ "modernc.org/sqlite"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrBranchLocked = errors.New("branch is locked")

//go:embed schema.sql
var schemaFS embed.FS

type Store struct {
	db      *sql.DB
	queries *db.Queries
	now     func() time.Time
}

type scanner interface {
	Scan(dest ...any) error
}

const envDBPath = "GIT_DIFF_DB_PATH"

func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}

	store := &Store{db: conn, queries: db.New(conn), now: time.Now}

	if err := store.Init(ctx); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return store, nil
}

func DefaultPath(home string) string {
	if override := os.Getenv(envDBPath); override != "" {
		return override
	}

	return filepath.Join(home, "Library", "Application Support", "git-diff", "reviews.sqlite3")
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	schema, err := schemaFS.ReadFile("schema.sql")

	if err != nil {
		return fmt.Errorf("read embedded sqlite schema: %w", err)
	}

	if err := s.dropLegacyProvisioningTables(ctx); err != nil {
		return fmt.Errorf("drop legacy provisioning tables: %w", err)
	}

	if err := s.addReviewSessionContextColumns(ctx); err != nil {
		return fmt.Errorf("add review_sessions context columns: %w", err)
	}

	if err := s.addCommentRangeColumns(ctx); err != nil {
		return fmt.Errorf("add comment range columns: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(schema)); err == nil {
		return nil
	}

	if err := s.resetSchema(ctx); err != nil {
		return fmt.Errorf("reset stale schema: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("initialize sqlite schema after reset: %w", err)
	}

	return nil
}

// addReviewSessionContextColumns is a one-shot migration that brings legacy
// databases up to the current schema by attaching the context_kind and
// context_sha columns to review_sessions. CREATE TABLE IF NOT EXISTS in
// schema.sql is a no-op for existing tables, so we need explicit ALTERs.
func (s *Store) addReviewSessionContextColumns(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, "PRAGMA table_info(review_sessions)")

	if err != nil {
		// Table does not exist yet — the schema apply below will create it
		// with the columns already in place.
		return nil
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
			return fmt.Errorf("scan column row: %w", err)
		}

		have[name] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate column rows: %w", err)
	}

	if len(have) == 0 {
		// Table not present — nothing to migrate.
		return nil
	}

	if !have["context_kind"] {
		if _, err := s.db.ExecContext(ctx, "ALTER TABLE review_sessions ADD COLUMN context_kind TEXT NOT NULL DEFAULT 'working'"); err != nil {
			return fmt.Errorf("add context_kind: %w", err)
		}
	}

	if !have["context_sha"] {
		if _, err := s.db.ExecContext(ctx, "ALTER TABLE review_sessions ADD COLUMN context_sha TEXT"); err != nil {
			return fmt.Errorf("add context_sha: %w", err)
		}
	}

	return nil
}

// addCommentRangeColumns brings legacy databases up to the current schema by
// attaching start_line_number / start_side columns to review_comments and
// pending_comments. Multi-line and cross-side comment ranges are stored
// through these fields; single-line comments leave both NULL.
func (s *Store) addCommentRangeColumns(ctx context.Context) error {
	for _, table := range []string{"review_comments", "pending_comments"} {
		rows, err := s.db.QueryContext(ctx, "PRAGMA table_info("+table+")")

		if err != nil {
			continue
		}

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
				rows.Close()

				return fmt.Errorf("scan %s column row: %w", table, err)
			}

			have[name] = true
		}

		rows.Close()

		if len(have) == 0 {
			continue
		}

		if !have["start_line_number"] {
			if _, err := s.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN start_line_number INTEGER"); err != nil {
				return fmt.Errorf("add %s.start_line_number: %w", table, err)
			}
		}

		if !have["start_side"] {
			if _, err := s.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN start_side TEXT"); err != nil {
				return fmt.Errorf("add %s.start_side: %w", table, err)
			}
		}
	}

	return nil
}

// dropLegacyProvisioningTables removes tables that used to back the macOS
// provisioning engine. They are no longer part of the schema; this lets
// existing user databases shed them on first launch after the upgrade.
func (s *Store) dropLegacyProvisioningTables(ctx context.Context) error {
	for _, table := range []string{"workflow_events", "workflow_runs"} {
		if _, err := s.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}

	return nil
}

func (s *Store) resetSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
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
		if _, err := s.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("drop table %q: %w", table, err)
		}
	}

	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}

func currentOSUsername() string {
	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name
	}

	if name := strings.TrimSpace(os.Getenv("USERNAME")); name != "" {
		return name
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Base(home)
	}

	return "user"
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func fromNull(value sql.NullString) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func nullableInt(value *int64) any {
	if value == nil {
		return nil
	}

	return *value
}
