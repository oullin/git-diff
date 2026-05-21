package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
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
