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

	"gorm.io/gorm"

	"github.com/gocanto/git-diff/internal/db"
)

type Session struct {
	ID         int64  `json:"id"`
	RawToken   string `json:"token"`
	UserID     int64  `json:"userId"`
	ExpiresAt  string `json:"expiresAt"`
	LastUsedAt string `json:"lastUsedAt"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type SessionRepo struct{}

func newSessionRepo() *SessionRepo {
	return &SessionRepo{}
}

func (r *SessionRepo) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (Session, error) {
	rawToken, err := generateToken(32)

	if err != nil {
		return Session{}, fmt.Errorf("generate session token: %w", err)
	}

	now := db.Now().UTC()
	created := now.Format(time.RFC3339Nano)
	expires := now.Add(ttl).Format(time.RFC3339Nano)
	stored := hashToken(rawToken)

	row := UserSessionRow{
		Token:      stored,
		UserID:     userID,
		ExpiresAt:  expires,
		LastUsedAt: created,
		CreatedAt:  created,
		UpdatedAt:  created,
	}

	if err := db.Conn().WithContext(ctx).Create(&row).Error; err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}

	return Session{
		ID:         row.ID,
		RawToken:   rawToken,
		UserID:     userID,
		ExpiresAt:  expires,
		LastUsedAt: created,
		CreatedAt:  created,
		UpdatedAt:  created,
	}, nil
}

func (r *SessionRepo) ResumeSession(ctx context.Context, users *UserRepo, rawToken string) (User, error) {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return User{}, ErrSessionNotFound
	}

	stored := hashToken(rawToken)

	var row UserSessionRow

	switch err := db.Conn().WithContext(ctx).Select("user_id", "expires_at").Where("token = ?", stored).Take(&row).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, sql.ErrNoRows):
		return User{}, ErrSessionNotFound
	case err != nil:
		return User{}, err
	}

	expiry, err := time.Parse(time.RFC3339Nano, row.ExpiresAt)

	if err == nil && db.Now().After(expiry) {
		_ = db.Conn().WithContext(ctx).Where("token = ?", stored).Delete(&UserSessionRow{}).Error

		return User{}, ErrSessionNotFound
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	if err := db.Conn().WithContext(ctx).
		Model(&UserSessionRow{}).
		Where("token = ?", stored).
		Updates(map[string]any{
			"last_used_at": now,
			"updated_at":   now,
		}).Error; err != nil {
		return User{}, err
	}

	return users.GetUserByID(ctx, row.UserID)
}

func (r *SessionRepo) DeleteSession(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return nil
	}

	return db.Conn().WithContext(ctx).Where("token = ?", hashToken(rawToken)).Delete(&UserSessionRow{}).Error
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
