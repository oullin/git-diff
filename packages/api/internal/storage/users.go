package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type User struct {
	ID           int64  `json:"id"`
	OSUsername   string `json:"osUsername"`
	DisplayName  string `json:"displayName"`
	HasPassword  bool   `json:"hasPassword"`
	CreatedAt    string `json:"createdAt"`
	LastLoginAt  string `json:"lastLoginAt,omitempty"`
	PasswordHash string `json:"-"`
}

type UserRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newUserRepo(conn *sql.DB, queries *db.Queries, clk *clock) *UserRepo {
	return &UserRepo{db: conn, queries: queries, clk: clk}
}

func (r *UserRepo) EnsureUser(ctx context.Context, osUsername string) (User, error) {
	osUsername = strings.TrimSpace(osUsername)

	if osUsername == "" {
		return User{}, errors.New("os username is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	if _, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO users (os_username, display_name, password_hash, created_at, last_login_at)
		VALUES (?, ?, '', ?, '')
	`, osUsername, osUsername, now); err != nil {
		return User{}, fmt.Errorf("seed user: %w", err)
	}

	return r.GetUserByOSUsername(ctx, osUsername)
}

func (r *UserRepo) GetUserByOSUsername(ctx context.Context, osUsername string) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, os_username, display_name, password_hash, created_at, last_login_at
		FROM users
		WHERE os_username = ?
	`, osUsername)

	return scanUser(row)
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, os_username, display_name, password_hash, created_at, last_login_at
		FROM users
		WHERE id = ?
	`, id)

	return scanUser(row)
}

func (r *UserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID)

	return err
}

func (r *UserRepo) TouchUserLogin(ctx context.Context, userID int64) error {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_login_at = ? WHERE id = ?`, now, userID)

	return err
}

func (r *UserRepo) WipeUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)

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
