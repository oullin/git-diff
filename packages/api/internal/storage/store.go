package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
	_ "modernc.org/sqlite"
)

// clock is shared across every repo so SetNow advances time everywhere.
type clock struct {
	now func() time.Time
}

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

	conn, err := sql.Open("sqlite", sqliteOpenDSN(path))

	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
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

func sqliteOpenDSN(path string) string {
	dsn := url.URL{Scheme: "file", Path: path}
	query := dsn.Query()
	query.Add("_pragma", "foreign_keys(1)")
	dsn.RawQuery = query.Encode()

	return dsn.String()
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

// SetNow swaps the clock used by every repository.
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
