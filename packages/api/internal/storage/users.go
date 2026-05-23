package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	db  *gorm.DB
	clk *clock
}

func newUserRepo(db *gorm.DB, clk *clock) *UserRepo {
	return &UserRepo{db: db, clk: clk}
}

func (r *UserRepo) EnsureUser(ctx context.Context, osUsername string) (User, error) {
	osUsername = strings.TrimSpace(osUsername)

	if osUsername == "" {
		return User{}, errors.New("os username is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	row := UserRow{
		OSUsername:  osUsername,
		DisplayName: osUsername,
		CreatedAt:   now,
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "os_username"}}, DoNothing: true}).
		Create(&row).Error; err != nil {
		return User{}, fmt.Errorf("seed user: %w", err)
	}

	return r.GetUserByOSUsername(ctx, osUsername)
}

func (r *UserRepo) GetUserByOSUsername(ctx context.Context, osUsername string) (User, error) {
	var row UserRow

	if err := r.db.WithContext(ctx).Where("os_username = ?", osUsername).Take(&row).Error; err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (User, error) {
	var row UserRow

	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&row).Error; err != nil {
		return User{}, err
	}

	return toUser(row), nil
}

func (r *UserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return r.db.WithContext(ctx).
		Model(&UserRow{}).
		Where("id = ?", userID).
		Update("password_hash", hash).Error
}

func (r *UserRepo) TouchUserLogin(ctx context.Context, userID int64) error {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	return r.db.WithContext(ctx).
		Model(&UserRow{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

func (r *UserRepo) WipeUser(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("id = ?", userID).Delete(&UserRow{}).Error
}

func toUser(row UserRow) User {
	return User{
		ID:           row.ID,
		OSUsername:   row.OSUsername,
		DisplayName:  row.DisplayName,
		HasPassword:  row.PasswordHash != "",
		CreatedAt:    row.CreatedAt,
		LastLoginAt:  row.LastLoginAt,
		PasswordHash: row.PasswordHash,
	}
}
