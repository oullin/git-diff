package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm/clause"

	"github.com/oullin/git-diff/internal/db"
)

type User struct {
	ID           int64  `json:"id"`
	OSUsername   string `json:"osUsername"`
	DisplayName  string `json:"displayName"`
	HasPassword  bool   `json:"hasPassword"`
	LastLoginAt  string `json:"lastLoginAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	PasswordHash string `json:"-"`
}

type UserRepo struct{}

func newUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) EnsureUser(ctx context.Context, osUsername string) (User, error) {
	osUsername = strings.TrimSpace(osUsername)

	if osUsername == "" {
		return User{}, errors.New("os username is required")
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	row := UserRow{
		OSUsername:  osUsername,
		DisplayName: osUsername,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := db.Conn().WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "os_username"}}, DoNothing: true}).
		Create(&row).Error; err != nil {
		return User{}, fmt.Errorf("seed user: %w", err)
	}

	return r.GetUserByOSUsername(ctx, osUsername)
}

func (r *UserRepo) GetUserByOSUsername(ctx context.Context, osUsername string) (User, error) {
	var row UserRow

	if err := db.Conn().WithContext(ctx).Where("os_username = ?", osUsername).Take(&row).Error; err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (User, error) {
	var row UserRow

	if err := db.Conn().WithContext(ctx).Where("id = ?", id).Take(&row).Error; err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (r *UserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return db.Conn().WithContext(ctx).
		Model(&UserRow{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash": hash,
			"updated_at":    db.Now().UTC().Format(time.RFC3339Nano),
		}).Error
}

func (r *UserRepo) TouchUserLogin(ctx context.Context, userID int64) error {
	now := db.Now().UTC().Format(time.RFC3339Nano)

	return db.Conn().WithContext(ctx).
		Model(&UserRow{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
}

func (r *UserRepo) WipeUser(ctx context.Context, userID int64) error {
	return db.Conn().WithContext(ctx).Where("id = ?", userID).Delete(&UserRow{}).Error
}

func toUser(row UserRow) User {
	return User{
		ID:           row.ID,
		OSUsername:   row.OSUsername,
		DisplayName:  row.DisplayName,
		HasPassword:  row.PasswordHash != "",
		LastLoginAt:  row.LastLoginAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		PasswordHash: row.PasswordHash,
	}
}
