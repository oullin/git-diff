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

	"github.com/oullin/git-diff/internal/db"
)

type Store struct {
	db *sql.DB

	Users           *UserRepo
	Sessions        *SessionRepo
	Reviews         *ReviewRepo
	ReviewEvents    *ReviewEventRepo
	Comments        *CommentRepo
	PendingComments *PendingCommentRepo
	Branches        *RepositoryBranchRepo
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

	db.Init(gdb)
	db.SetNow(time.Now)

	reviewEvents := newReviewEventRepo()

	store := &Store{
		db:              conn,
		Users:           newUserRepo(),
		Sessions:        newSessionRepo(),
		Reviews:         newReviewRepo(reviewEvents),
		ReviewEvents:    reviewEvents,
		Comments:        newCommentRepo(reviewEvents),
		PendingComments: newPendingCommentRepo(),
		Branches:        newRepositoryBranchRepo(),
		Repos:           newRepoRepo(),
		Collaborators:   newCollaboratorRepo(),
		Preferences:     newPreferenceRepo(),
		Walkthroughs:    newWalkthroughRepo(),
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

// SetNow swaps the clock used by every repository.
func (s *Store) SetNow(fn func() time.Time) {
	db.SetNow(fn)
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
