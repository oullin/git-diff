package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
	_ "modernc.org/sqlite"
)

const DefaultTheme = "light"
const DefaultDiffViewMode = "split"

const (
	PrefKeyTheme              = "theme"
	PrefKeyDiffViewMode       = "diff.viewMode"
	PrefKeyDiffHideWhitespace = "diff.hideWhitespace"
	PrefKeyLastRepoRoot       = "repo.lastRoot"
	PrefKeyPanelLeftWidth     = "panel.left.width"
	PrefKeyPanelFileTreeWidth = "panel.fileTree.width"
	PrefKeyPanelRightWidth    = "panel.right.width"
)

var ErrSessionNotFound = errors.New("session not found")

//go:embed schema.sql
var schemaFS embed.FS

type Store struct {
	db      *sql.DB
	queries *db.Queries
	now     func() time.Time
}

type UIPreferences struct {
	Values    map[string]string `json:"values"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

type User struct {
	ID           int64  `json:"id"`
	OSUsername   string `json:"osUsername"`
	DisplayName  string `json:"displayName"`
	HasPassword  bool   `json:"hasPassword"`
	CreatedAt    string `json:"createdAt"`
	LastLoginAt  string `json:"lastLoginAt,omitempty"`
	PasswordHash string `json:"-"`
}

type Session struct {
	RawToken   string `json:"token"`
	UserID     int64  `json:"userId"`
	CreatedAt  string `json:"createdAt"`
	ExpiresAt  string `json:"expiresAt"`
	LastUsedAt string `json:"lastUsedAt"`
}

type ReviewSessionStart struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha"`
}

type ReviewSession struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	UserID       int64  `json:"userId"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	StartedAt    string `json:"startedAt"`
	CompletedAt  string `json:"completedAt,omitempty"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha,omitempty"`
}

type ReviewEventInput struct {
	Type     string `json:"type"`
	FilePath string `json:"filePath"`
	Message  string `json:"message"`
	Metadata string `json:"metadata"`
}

type ReviewEvent struct {
	ID        int64  `json:"id"`
	ReviewID  string `json:"reviewId"`
	Type      string `json:"type"`
	FilePath  string `json:"filePath,omitempty"`
	Message   string `json:"message,omitempty"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"createdAt"`
}

type ReviewCommentInput struct {
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
}

type ReviewComment struct {
	ID          string `json:"id"`
	ReviewID    string `json:"reviewId"`
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DeletedAt   string `json:"deletedAt,omitempty"`
}

type ReviewDetail struct {
	Review   ReviewSession   `json:"review"`
	Events   []ReviewEvent   `json:"events"`
	Comments []ReviewComment `json:"comments"`
}

type Repository struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	OwnerID      int64  `json:"ownerId"`
	Role         string `json:"role"`
	AddedAt      string `json:"addedAt"`
	LastOpenedAt string `json:"lastOpenedAt,omitempty"`
}

type RepositoryCollaborator struct {
	UserID      int64  `json:"userId"`
	OSUsername  string `json:"osUsername"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	GrantedAt   string `json:"grantedAt"`
}

type Branch struct {
	Name       string `json:"name"`
	Locked     bool   `json:"locked"`
	LockedBy   int64  `json:"lockedBy,omitempty"`
	LockedAt   string `json:"lockedAt,omitempty"`
	LastSeenAt string `json:"lastSeenAt"`
}

type scanner interface {
	Scan(dest ...any) error
}

const (
	RepoRoleOwner = "owner"
	RepoRoleWrite = "write"
	RepoRoleRead  = "read"
)

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

func (s *Store) GetUIPreferences(ctx context.Context, userID int64) (UIPreferences, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT key, value, updated_at
		FROM ui_preferences
		WHERE user_id = ?
	`, userID)

	if err != nil {
		return UIPreferences{}, err
	}

	defer rows.Close()

	prefs := UIPreferences{Values: map[string]string{}}

	for rows.Next() {
		var (
			key       string
			value     string
			updatedAt string
		)

		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			return UIPreferences{}, err
		}

		prefs.Values[key] = value

		if updatedAt > prefs.UpdatedAt {
			prefs.UpdatedAt = updatedAt
		}
	}

	return prefs, rows.Err()
}

func (s *Store) SaveUIPreferences(ctx context.Context, userID int64, patch map[string]string) (UIPreferences, error) {
	if len(patch) == 0 {
		return s.GetUIPreferences(ctx, userID)
	}

	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return UIPreferences{}, err
	}

	defer tx.Rollback()

	updatedAt := s.now().UTC().Format(time.RFC3339Nano)

	for key, value := range patch {
		key = strings.TrimSpace(key)

		if key == "" {
			continue
		}

		if value == "" {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM ui_preferences WHERE user_id = ? AND key = ?
			`, userID, key); err != nil {
				return UIPreferences{}, err
			}

			continue
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO ui_preferences (user_id, key, value, updated_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, key) DO UPDATE SET
				value = excluded.value,
				updated_at = excluded.updated_at
		`, userID, key, value, updatedAt); err != nil {
			return UIPreferences{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return UIPreferences{}, err
	}

	return s.GetUIPreferences(ctx, userID)
}

func (s *Store) EnsureUser(ctx context.Context, osUsername string) (User, error) {
	osUsername = strings.TrimSpace(osUsername)

	if osUsername == "" {
		return User{}, errors.New("os username is required")
	}

	now := s.now().UTC().Format(time.RFC3339Nano)

	if _, err := s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO users (os_username, display_name, password_hash, created_at, last_login_at)
		VALUES (?, ?, '', ?, '')
	`, osUsername, osUsername, now); err != nil {
		return User{}, fmt.Errorf("seed user: %w", err)
	}

	return s.GetUserByOSUsername(ctx, osUsername)
}

func (s *Store) GetUserByOSUsername(ctx context.Context, osUsername string) (User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, os_username, display_name, password_hash, created_at, last_login_at
		FROM users
		WHERE os_username = ?
	`, osUsername)

	return scanUser(row)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, os_username, display_name, password_hash, created_at, last_login_at
		FROM users
		WHERE id = ?
	`, id)

	return scanUser(row)
}

func (s *Store) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID)

	return err
}

func (s *Store) TouchUserLogin(ctx context.Context, userID int64) error {
	now := s.now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `UPDATE users SET last_login_at = ? WHERE id = ?`, now, userID)

	return err
}

func (s *Store) WipeUser(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)

	return err
}

func (s *Store) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (Session, error) {
	rawToken, err := generateToken(32)

	if err != nil {
		return Session{}, fmt.Errorf("generate session token: %w", err)
	}

	now := s.now().UTC()
	created := now.Format(time.RFC3339Nano)
	expires := now.Add(ttl).Format(time.RFC3339Nano)
	stored := hashToken(rawToken)

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO user_sessions (token, user_id, created_at, expires_at, last_used_at)
		VALUES (?, ?, ?, ?, ?)
	`, stored, userID, created, expires, created); err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}

	return Session{
		RawToken:   rawToken,
		UserID:     userID,
		CreatedAt:  created,
		ExpiresAt:  expires,
		LastUsedAt: created,
	}, nil
}

func (s *Store) ResumeSession(ctx context.Context, rawToken string) (User, error) {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return User{}, ErrSessionNotFound
	}

	stored := hashToken(rawToken)
	row := s.db.QueryRowContext(ctx, `
		SELECT user_id, expires_at
		FROM user_sessions
		WHERE token = ?
	`, stored)

	var (
		userID    int64
		expiresAt string
	)

	switch err := row.Scan(&userID, &expiresAt); {
	case errors.Is(err, sql.ErrNoRows):
		return User{}, ErrSessionNotFound
	case err != nil:
		return User{}, err
	}

	expiry, err := time.Parse(time.RFC3339Nano, expiresAt)

	if err == nil && s.now().After(expiry) {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token = ?`, stored)

		return User{}, ErrSessionNotFound
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE user_sessions SET last_used_at = ? WHERE token = ?
	`, s.now().UTC().Format(time.RFC3339Nano), stored); err != nil {
		return User{}, err
	}

	return s.GetUserByID(ctx, userID)
}

func (s *Store) DeleteSession(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return nil
	}

	_, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token = ?`, hashToken(rawToken))

	return err
}

func scanUser(row scanner) (User, error) {
	var (
		user        User
		lastLoginAt string
	)

	if err := row.Scan(
		&user.ID,
		&user.OSUsername,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&lastLoginAt,
	); err != nil {
		return User{}, err
	}

	user.LastLoginAt = lastLoginAt
	user.HasPassword = user.PasswordHash != ""

	return user, nil
}

func generateToken(size int) (string, error) {
	buf := make([]byte, size)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))

	return hex.EncodeToString(sum[:])
}

func (s *Store) CreateReview(ctx context.Context, userID int64, review ReviewSessionStart) (ReviewSession, error) {
	if userID == 0 {
		return ReviewSession{}, errors.New("user id is required")
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
	title := review.Title

	if title == "" {
		title = "Local review"
	}

	contextKind := review.ContextKind

	if contextKind == "" {
		contextKind = "working"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_sessions (
			id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, context_kind, context_sha
		) VALUES (?, ?, ?, ?, ?, 'open', ?, ?, ?, ?, ?, ?, ?, ?)
	`, review.ID, review.RepoRoot, userID, review.Branch, review.HeadSHA, title, review.Summary, review.FilesChanged, review.Additions, review.Deletions, now, contextKind, nullString(review.ContextSHA))

	if err != nil {
		return ReviewSession{}, err
	}

	if _, err := s.AddReviewEvent(ctx, review.ID, ReviewEventInput{
		Type:    "review_started",
		Message: title,
	}); err != nil {
		return ReviewSession{}, err
	}

	return ReviewSession{
		ID:           review.ID,
		RepoRoot:     review.RepoRoot,
		UserID:       userID,
		Branch:       review.Branch,
		HeadSHA:      review.HeadSHA,
		Status:       "open",
		Title:        title,
		Summary:      review.Summary,
		FilesChanged: review.FilesChanged,
		Additions:    review.Additions,
		Deletions:    review.Deletions,
		StartedAt:    now,
		ContextKind:  contextKind,
		ContextSHA:   review.ContextSHA,
	}, nil
}

func (s *Store) ListReviews(ctx context.Context, userID int64, limit int64) ([]ReviewSession, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at,
			context_kind, context_sha
		FROM review_sessions
		WHERE user_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, userID, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reviews := []ReviewSession{}

	for rows.Next() {
		review, err := scanReview(rows)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

func (s *Store) ReviewDetail(ctx context.Context, id string) (ReviewDetail, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, repo_root, user_id, branch, head_sha, status, title, summary,
			files_changed, additions, deletions, started_at, completed_at,
			context_kind, context_sha
		FROM review_sessions
		WHERE id = ?
	`, id)
	review, err := scanReview(row)

	if err != nil {
		return ReviewDetail{}, err
	}

	events, err := s.ListReviewEvents(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	comments, err := s.ListReviewComments(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	return ReviewDetail{Review: review, Events: events, Comments: comments}, nil
}

func (s *Store) AddReviewEvent(ctx context.Context, reviewID string, input ReviewEventInput) (ReviewEvent, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)
	metadata := input.Metadata

	if metadata == "" {
		metadata = "{}"
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO review_events (review_id, event_type, file_path, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reviewID, input.Type, input.FilePath, input.Message, metadata, now)

	if err != nil {
		return ReviewEvent{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return ReviewEvent{}, err
	}

	return ReviewEvent{ID: id, ReviewID: reviewID, Type: input.Type, FilePath: input.FilePath, Message: input.Message, Metadata: metadata, CreatedAt: now}, nil
}

func (s *Store) ListReviewEvents(ctx context.Context, reviewID string) ([]ReviewEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, review_id, event_type, file_path, message, metadata, created_at
		FROM review_events
		WHERE review_id = ?
		ORDER BY created_at ASC, id ASC
	`, reviewID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	events := []ReviewEvent{}

	for rows.Next() {
		var event ReviewEvent

		if err := rows.Scan(&event.ID, &event.ReviewID, &event.Type, &event.FilePath, &event.Message, &event.Metadata, &event.CreatedAt); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

func (s *Store) CreateReviewComment(ctx context.Context, reviewID string, commentID string, input ReviewCommentInput) (ReviewComment, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)

	if input.AuthorLabel == "" {
		input.AuthorLabel = "You"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_comments (
			id, review_id, file_path, diff_section, side, line_number,
			author_label, body_html, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, commentID, reviewID, input.FilePath, input.DiffSection, input.Side, input.LineNumber, input.AuthorLabel, input.BodyHTML, now, now)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_added", FilePath: input.FilePath, Message: fmt.Sprintf("Commented on line %d", input.LineNumber)})

	return ReviewComment{
		ID: commentID, ReviewID: reviewID, FilePath: input.FilePath, DiffSection: input.DiffSection,
		Side: input.Side, LineNumber: input.LineNumber, AuthorLabel: input.AuthorLabel,
		BodyHTML: input.BodyHTML, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Store) UpdateReviewComment(ctx context.Context, reviewID string, commentID string, bodyHTML string) (ReviewComment, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		UPDATE review_comments
		SET body_html = ?, updated_at = ?, deleted_at = NULL
		WHERE id = ? AND review_id = ?
	`, bodyHTML, now, commentID, reviewID)

	if err != nil {
		return ReviewComment{}, err
	}

	comment, err := s.GetReviewComment(ctx, reviewID, commentID)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_edited", FilePath: comment.FilePath, Message: fmt.Sprintf("Edited comment on line %d", comment.LineNumber)})

	return comment, nil
}

func (s *Store) DeleteReviewComment(ctx context.Context, reviewID string, commentID string) error {
	now := s.now().UTC().Format(time.RFC3339Nano)
	comment, _ := s.GetReviewComment(ctx, reviewID, commentID)
	_, err := s.db.ExecContext(ctx, `
		UPDATE review_comments
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND review_id = ?
	`, now, now, commentID, reviewID)

	if err != nil {
		return err
	}

	if comment.ID != "" {
		_, _ = s.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_deleted", FilePath: comment.FilePath, Message: fmt.Sprintf("Deleted comment on line %d", comment.LineNumber)})
	}

	return nil
}

func (s *Store) GetReviewComment(ctx context.Context, reviewID string, commentID string) (ReviewComment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE id = ? AND review_id = ?
	`, commentID, reviewID)

	return scanComment(row)
}

func (s *Store) ListReviewComments(ctx context.Context, reviewID string) ([]ReviewComment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE review_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, reviewID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []ReviewComment{}

	for rows.Next() {
		comment, err := scanComment(rows)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func (s *Store) ListRepositoriesForUser(ctx context.Context, userID int64) ([]Repository, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			CASE WHEN r.owner_id = ?1 THEN 'owner' ELSE ru.role END AS role
		FROM repositories r
		LEFT JOIN repository_users ru ON ru.repo_path = r.path AND ru.user_id = ?1
		WHERE r.owner_id = ?1 OR ru.user_id = ?1
		ORDER BY COALESCE(r.last_opened_at, r.added_at) DESC
	`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := []Repository{}

	for rows.Next() {
		var (
			repo         Repository
			lastOpenedAt sql.NullString
			role         sql.NullString
		)

		if err := rows.Scan(&repo.Path, &repo.Name, &repo.OwnerID, &repo.AddedAt, &lastOpenedAt, &role); err != nil {
			return nil, err
		}

		repo.LastOpenedAt = fromNull(lastOpenedAt)
		repo.Role = fromNull(role)
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

func (s *Store) GetRepository(ctx context.Context, userID int64, path string) (Repository, error) {
	if userID == 0 {
		return Repository{}, errors.New("user id is required")
	}

	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			CASE WHEN r.owner_id = ?1 THEN 'owner' ELSE ru.role END AS role
		FROM repositories r
		LEFT JOIN repository_users ru ON ru.repo_path = r.path AND ru.user_id = ?1
		WHERE r.path = ?2 AND (r.owner_id = ?1 OR ru.user_id = ?1)
	`, userID, path)

	var (
		repo         Repository
		lastOpenedAt sql.NullString
		role         sql.NullString
	)

	if err := row.Scan(&repo.Path, &repo.Name, &repo.OwnerID, &repo.AddedAt, &lastOpenedAt, &role); err != nil {
		return Repository{}, err
	}

	repo.LastOpenedAt = fromNull(lastOpenedAt)
	repo.Role = fromNull(role)

	return repo, nil
}

func (s *Store) UpsertRepository(ctx context.Context, ownerID int64, path string, name string) (Repository, error) {
	if ownerID == 0 {
		return Repository{}, errors.New("owner id is required")
	}

	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	if name == "" {
		name = filepath.Base(path)
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO repositories (path, name, owner_id, added_at, last_opened_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			last_opened_at = excluded.last_opened_at
	`, path, name, ownerID, now, now)

	if err != nil {
		return Repository{}, err
	}

	return s.GetRepository(ctx, ownerID, path)
}

func (s *Store) RemoveRepository(ctx context.Context, userID int64, path string) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	if path == "" {
		return errors.New("repository path is required")
	}

	result, err := s.db.ExecContext(ctx, `
		DELETE FROM repositories WHERE path = ? AND owner_id = ?
	`, path, userID)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrRepositoryNotOwned
	}

	return nil
}

var ErrRepositoryNotOwned = errors.New("repository not found or not owned by user")

func scanReview(row scanner) (ReviewSession, error) {
	var review ReviewSession

	var (
		completedAt sql.NullString
		contextSHA  sql.NullString
	)

	if err := row.Scan(
		&review.ID,
		&review.RepoRoot,
		&review.UserID,
		&review.Branch,
		&review.HeadSHA,
		&review.Status,
		&review.Title,
		&review.Summary,
		&review.FilesChanged,
		&review.Additions,
		&review.Deletions,
		&review.StartedAt,
		&completedAt,
		&review.ContextKind,
		&contextSHA,
	); err != nil {
		return ReviewSession{}, err
	}

	review.CompletedAt = fromNull(completedAt)
	review.ContextSHA = fromNull(contextSHA)

	return review, nil
}

func scanComment(row scanner) (ReviewComment, error) {
	var comment ReviewComment

	var deletedAt sql.NullString

	if err := row.Scan(
		&comment.ID,
		&comment.ReviewID,
		&comment.FilePath,
		&comment.DiffSection,
		&comment.Side,
		&comment.LineNumber,
		&comment.AuthorLabel,
		&comment.BodyHTML,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&deletedAt,
	); err != nil {
		return ReviewComment{}, err
	}

	comment.DeletedAt = fromNull(deletedAt)

	return comment, nil
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
