package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type Session struct {
	RawToken   string `json:"token"`
	UserID     int64  `json:"userId"`
	CreatedAt  string `json:"createdAt"`
	ExpiresAt  string `json:"expiresAt"`
	LastUsedAt string `json:"lastUsedAt"`
}

type SessionRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newSessionRepo(conn *sql.DB, queries *db.Queries, clk *clock) *SessionRepo {
	return &SessionRepo{db: conn, queries: queries, clk: clk}
}

func (r *SessionRepo) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (Session, error) {
	rawToken, err := generateToken(32)

	if err != nil {
		return Session{}, fmt.Errorf("generate session token: %w", err)
	}

	now := r.clk.now().UTC()
	created := now.Format(time.RFC3339Nano)
	expires := now.Add(ttl).Format(time.RFC3339Nano)
	stored := hashToken(rawToken)

	if _, err := r.db.ExecContext(ctx, `
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

// ResumeSession verifies the token, updates last_used_at, and returns the
// associated user. The UserRepo is provided so callers wire the lookup
// explicitly instead of through a god-object Store.
func (r *SessionRepo) ResumeSession(ctx context.Context, users *UserRepo, rawToken string) (User, error) {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return User{}, ErrSessionNotFound
	}

	stored := hashToken(rawToken)
	row := r.db.QueryRowContext(ctx, `
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

	if err == nil && r.clk.now().After(expiry) {
		_, _ = r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token = ?`, stored)

		return User{}, ErrSessionNotFound
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions SET last_used_at = ? WHERE token = ?
	`, r.clk.now().UTC().Format(time.RFC3339Nano), stored); err != nil {
		return User{}, err
	}

	return users.GetUserByID(ctx, userID)
}

func (r *SessionRepo) DeleteSession(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token = ?`, hashToken(rawToken))

	return err
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
