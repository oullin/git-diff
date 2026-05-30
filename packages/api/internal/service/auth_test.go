package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oullin/git-diff/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func newAuthService(t *testing.T, store *storage.Store, cfg AuthConfig) *AuthService {
	t.Helper()

	return NewAuthService(store.Users, store.Sessions, cfg)
}

func TestNewAuthServiceAppliesDefaults(t *testing.T) {
	store := newTestStore(t)
	svc := newAuthService(t, store, AuthConfig{})

	if svc.bcryptCost != bcrypt.DefaultCost {
		t.Fatalf("expected bcrypt default cost, got %d", svc.bcryptCost)
	}

	if svc.sessionTTL != 90*24*time.Hour {
		t.Fatalf("expected 90d default TTL, got %v", svc.sessionTTL)
	}

	if svc.minPasswordLength != 6 {
		t.Fatalf("expected default min password length 6, got %d", svc.minPasswordLength)
	}
}

func TestSetupRejectsShortPassword(t *testing.T) {
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost, MinPasswordLength: 6})

	if _, _, err := svc.Setup(context.Background(), "alice", "abc"); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestSetupCreatesPasswordAndSession(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	user, session, err := svc.Setup(ctx, "alice", "secret123")

	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if !user.HasPassword {
		t.Fatalf("expected HasPassword=true")
	}

	if session.RawToken == "" {
		t.Fatalf("expected raw token issued by Setup")
	}
}

func TestSetupRejectsAlreadyConfigured(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); !errors.Is(err, ErrPasswordAlreadySet) {
		t.Fatalf("expected ErrPasswordAlreadySet, got %v", err)
	}
}

func TestLoginRequiresPasswordSet(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Login(ctx, "alice", "anything", false); !errors.Is(err, ErrPasswordNotSet) {
		t.Fatalf("expected ErrPasswordNotSet, got %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, _, err := svc.Login(ctx, "alice", "wrong", false); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLoginRememberFalseDoesNotMintSession(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, session, err := svc.Login(ctx, "alice", "secret123", false)

	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if session != nil {
		t.Fatalf("expected no session when remember=false, got %+v", session)
	}
}

func TestLoginRememberTrueMintsSession(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, session, err := svc.Login(ctx, "alice", "secret123", true)

	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if session == nil || session.RawToken == "" {
		t.Fatalf("expected session with raw token, got %+v", session)
	}
}

func TestResumePropagatesSessionNotFound(t *testing.T) {
	store := newTestStore(t)
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, err := svc.Resume(context.Background(), "unknown-token"); !errors.Is(err, storage.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestLogoutNoOpForBlankToken(t *testing.T) {
	store := newTestStore(t)
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatalf("blank logout must be a no-op, got %v", err)
	}

	if err := svc.Logout(context.Background(), "   "); err != nil {
		t.Fatalf("whitespace logout must be a no-op, got %v", err)
	}
}

func TestWipeRefusesNonActiveUser(t *testing.T) {
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if err := svc.Wipe(context.Background(), "alice", "bob"); !errors.Is(err, ErrCannotWipeOtherUser) {
		t.Fatalf("expected ErrCannotWipeOtherUser, got %v", err)
	}
}

func TestWipeReseedsUser(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if _, _, err := svc.Setup(ctx, "alice", "secret123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := svc.Wipe(ctx, "alice", "alice"); err != nil {
		t.Fatalf("wipe: %v", err)
	}

	user, err := store.Users.GetUserByOSUsername(ctx, "alice")

	if err != nil {
		t.Fatalf("get reseeded user: %v", err)
	}

	if user.HasPassword {
		t.Fatalf("expected fresh user without password after wipe, got HasPassword=true")
	}
}

func TestWipeFallsBackToActiveUserWhenRequestedIsBlank(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	seedUser(t, store, "alice")
	svc := newAuthService(t, store, AuthConfig{BcryptCost: bcrypt.MinCost})

	if err := svc.Wipe(ctx, "alice", "   "); err != nil {
		t.Fatalf("expected blank target to fall back to active user, got %v", err)
	}
}
