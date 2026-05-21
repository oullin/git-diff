// Package service hosts orchestration logic that lives between the HTTP
// handlers and the storage layer. Each service exposes a small interface
// expressing one domain's use cases (auth, review, repository, ...) and
// encapsulates the cross-cutting concerns (password hashing, session TTL,
// transactional bookkeeping) that have no business in transport code.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// AuthService owns the user-authentication use cases. Construction time
// captures the bcrypt cost, session TTL, and minimum password length so the
// HTTP layer never has to reach for these knobs.
type AuthService struct {
	store             *storage.Store
	bcryptCost        int
	sessionTTL        time.Duration
	minPasswordLength int
}

type AuthConfig struct {
	BcryptCost        int
	SessionTTL        time.Duration
	MinPasswordLength int
}

func NewAuthService(store *storage.Store, cfg AuthConfig) *AuthService {
	if cfg.BcryptCost == 0 {
		cfg.BcryptCost = bcrypt.DefaultCost
	}

	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 90 * 24 * time.Hour
	}

	if cfg.MinPasswordLength == 0 {
		cfg.MinPasswordLength = 6
	}

	return &AuthService{
		store:             store,
		bcryptCost:        cfg.BcryptCost,
		sessionTTL:        cfg.SessionTTL,
		minPasswordLength: cfg.MinPasswordLength,
	}
}

// Domain errors signal expected business outcomes that the handler layer
// translates into specific HTTP status codes.
var (
	ErrPasswordTooShort    = errors.New("password too short")
	ErrPasswordAlreadySet  = errors.New("password already set")
	ErrPasswordNotSet      = errors.New("password not set")
	ErrInvalidPassword     = errors.New("incorrect password")
	ErrCannotWipeOtherUser = errors.New("can only wipe the active OS user")
)

// State returns the user record used to populate authStateResponse.
func (s *AuthService) State(ctx context.Context, osUsername string) (storage.User, error) {
	return s.store.GetUserByOSUsername(ctx, osUsername)
}

// Setup hashes the supplied password, stores it for the active OS user, and
// creates the initial remembered session.
func (s *AuthService) Setup(
	ctx context.Context,
	osUsername, password string,
) (storage.User, storage.Session, error) {
	if len(strings.TrimSpace(password)) < s.minPasswordLength {
		return storage.User{}, storage.Session{}, ErrPasswordTooShort
	}

	user, err := s.store.GetUserByOSUsername(ctx, osUsername)

	if err != nil {
		return storage.User{}, storage.Session{}, fmt.Errorf("read user: %w", err)
	}

	if user.HasPassword {
		return storage.User{}, storage.Session{}, ErrPasswordAlreadySet
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)

	if err != nil {
		return storage.User{}, storage.Session{}, fmt.Errorf("hash password: %w", err)
	}

	if err := s.store.SetPasswordHash(ctx, user.ID, string(hash)); err != nil {
		return storage.User{}, storage.Session{}, fmt.Errorf("save password hash: %w", err)
	}

	session, err := s.store.CreateSession(ctx, user.ID, s.sessionTTL)

	if err != nil {
		return storage.User{}, storage.Session{}, fmt.Errorf("create session: %w", err)
	}

	_ = s.store.TouchUserLogin(ctx, user.ID)

	user.HasPassword = true

	return user, session, nil
}

// Login validates the supplied password. If remember is true a persistent
// session is created and returned; otherwise the session pointer is nil and
// callers should record only the in-memory user id.
func (s *AuthService) Login(
	ctx context.Context,
	osUsername, password string,
	remember bool,
) (storage.User, *storage.Session, error) {
	user, err := s.store.GetUserByOSUsername(ctx, osUsername)

	if err != nil {
		return storage.User{}, nil, fmt.Errorf("read user: %w", err)
	}

	if !user.HasPassword {
		return storage.User{}, nil, ErrPasswordNotSet
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return storage.User{}, nil, ErrInvalidPassword
	}

	_ = s.store.TouchUserLogin(ctx, user.ID)

	if !remember {
		return user, nil, nil
	}

	session, err := s.store.CreateSession(ctx, user.ID, s.sessionTTL)

	if err != nil {
		return storage.User{}, nil, fmt.Errorf("create session: %w", err)
	}

	return user, &session, nil
}

// Resume restores the user behind a stored session token. ErrSessionNotFound
// is propagated from the store untouched so handlers can map it to 401.
func (s *AuthService) Resume(ctx context.Context, token string) (storage.User, error) {
	return s.store.ResumeSession(ctx, token)
}

// Logout drops the session row for the given token. A blank token is a no-op
// so handlers can call this unconditionally.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}

	return s.store.DeleteSession(ctx, token)
}

// Wipe deletes the user and re-seeds a fresh, password-less account for the
// same OS username. Only the active OS user may wipe their own account.
func (s *AuthService) Wipe(ctx context.Context, activeOSUsername, requestedOSUsername string) error {
	target := strings.TrimSpace(requestedOSUsername)

	if target == "" {
		target = activeOSUsername
	}

	if target != activeOSUsername {
		return ErrCannotWipeOtherUser
	}

	user, err := s.store.GetUserByOSUsername(ctx, target)

	if err != nil {
		return fmt.Errorf("read user: %w", err)
	}

	if err := s.store.WipeUser(ctx, user.ID); err != nil {
		return fmt.Errorf("wipe user: %w", err)
	}

	if _, err := s.store.EnsureUser(ctx, target); err != nil {
		return fmt.Errorf("reseed user: %w", err)
	}

	return nil
}
