package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// clock is shared across every repo so SetNow advances time everywhere.
type clock struct {
	now func() time.Time
}

type Store struct {
	db  *sql.DB
	gdb *gorm.DB
	clk *clock

	Users           *UserRepo
	Sessions        *SessionRepo
	Reviews         *ReviewRepo
	ReviewEvents    *ReviewEventRepo
	Comments        *CommentRepo
	PendingComments *PendingCommentRepo
	Branches        *BranchRepo
	Repos           *RepoRepo
	Collaborators   *CollaboratorRepo
	Preferences     *PreferenceRepo
	Walkthroughs    *WalkthroughRepo
}

var ErrSessionNotFound = errors.New("session not found")
var ErrBranchLocked = errors.New("branch is locked")

const envDBPath = "GIT_DIFF_DB_PATH"

func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	gdb, err := gorm.Open(sqlite.Open(sqliteOpenDSN(path)), &gorm.Config{
		Logger: gormlogger.New(log.New(os.Stderr, "[gorm] ", log.LstdFlags), gormlogger.Config{
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	conn, err := gdb.DB()

	if err != nil {
		return nil, fmt.Errorf("unwrap *sql.DB from gorm: %w", err)
	}

	clk := &clock{now: time.Now}
	reviewEvents := newReviewEventRepo(gdb, clk)

	store := &Store{
		db:              conn,
		gdb:             gdb,
		clk:             clk,
		Users:           newUserRepo(gdb, clk),
		Sessions:        newSessionRepo(gdb, clk),
		Reviews:         newReviewRepo(gdb, clk, reviewEvents),
		ReviewEvents:    reviewEvents,
		Comments:        newCommentRepo(gdb, clk, reviewEvents),
		PendingComments: newPendingCommentRepo(gdb, clk),
		Branches:        newBranchRepo(gdb, clk),
		Repos:           newRepoRepo(gdb, clk),
		Collaborators:   newCollaboratorRepo(gdb, clk),
		Preferences:     newPreferenceRepo(gdb, clk),
		Walkthroughs:    newWalkthroughRepo(gdb, clk),
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
	query.Add("_foreign_keys", "on")
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

// DB exposes the underlying *gorm.DB for repos and advanced chained queries.
func (s *Store) DB() *gorm.DB { return s.gdb }

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
