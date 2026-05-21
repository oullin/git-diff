package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
	_ "modernc.org/sqlite"
)

// clock is a small mutable holder for the time source. Repositories share a
// pointer to one clock so a test can advance time on the Store and every
// repo observes it through r.clk.now().
type clock struct {
	now func() time.Time
}

// Store is the composition root for the SQLite-backed storage layer. It owns
// the database connection, the sqlc-generated queries, and a shared clock,
// and exposes one repository per domain. Services depend on the specific
// repositories they need rather than on the Store itself.
type Store struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock

	Users           *UserRepo
	Sessions        *SessionRepo
	Reviews         *ReviewRepo
	Comments        *CommentRepo
	PendingComments *PendingCommentRepo
	Branches        *BranchRepo
	Repos           *RepoRepo
	Preferences     *PreferenceRepo
	Walkthroughs    *WalkthroughRepo
}

type scanner interface {
	Scan(dest ...any) error
}

var ErrSessionNotFound = errors.New("session not found")
var ErrBranchLocked = errors.New("branch is locked")

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

	clk := &clock{now: time.Now}
	queries := db.New(conn)
	reviews := newReviewRepo(conn, queries, clk)

	store := &Store{
		db:              conn,
		queries:         queries,
		clk:             clk,
		Users:           newUserRepo(conn, queries, clk),
		Sessions:        newSessionRepo(conn, queries, clk),
		Reviews:         reviews,
		Comments:        newCommentRepo(conn, queries, clk, reviews),
		PendingComments: newPendingCommentRepo(conn, queries, clk),
		Branches:        newBranchRepo(conn, queries, clk),
		Repos:           newRepoRepo(conn, queries, clk),
		Preferences:     newPreferenceRepo(conn, queries, clk),
		Walkthroughs:    newWalkthroughRepo(conn, queries, clk),
	}

	if err := NewMigrator(conn).Run(ctx); err != nil {
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

// SetNow swaps the clock used by every repository. Tests use this to
// advance time deterministically.
func (s *Store) SetNow(fn func() time.Time) {
	s.clk.now = fn
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
