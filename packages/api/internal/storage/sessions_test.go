package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateSessionReturnsRawTokenAndHashesItOnDisk(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	session, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if session.RawToken == "" {
		t.Fatalf("expected raw token")
	}

	// Same RawToken hashed twice must collide; raw token must not be stored.
	stored := hashToken(session.RawToken)

	if stored == session.RawToken {
		t.Fatalf("raw token must not equal stored hash")
	}

	// Direct table inspection: the row's Token column should equal stored hash.
	var row UserSessionRow

	if err := store.db.QueryRowContext(ctx, "SELECT token FROM user_sessions WHERE id = ?", session.ID).
		Scan(&row.Token); err != nil {
		t.Fatalf("inspect row: %v", err)
	}

	if row.Token != stored {
		t.Fatalf("stored token %q != hash(raw) %q", row.Token, stored)
	}
}

func TestResumeSessionRejectsBlankToken(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Sessions.ResumeSession(context.Background(), store.Users, "   ")

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for blank token, got %v", err)
	}
}

func TestResumeSessionRejectsUnknownToken(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Sessions.ResumeSession(context.Background(), store.Users, "never-existed")

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestResumeSessionUpdatesLastUsedAt(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	session, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	initialLastUsed := session.LastUsedAt

	// Advance the clock so the LastUsedAt update is observable.
	store.SetNow(func() time.Time { return time.Now().Add(5 * time.Minute) })

	if _, err := store.Sessions.ResumeSession(ctx, store.Users, session.RawToken); err != nil {
		t.Fatalf("resume: %v", err)
	}

	store.SetNow(time.Now)

	var lastUsed string

	if err := store.db.QueryRowContext(ctx, "SELECT last_used_at FROM user_sessions WHERE id = ?", session.ID).
		Scan(&lastUsed); err != nil {
		t.Fatalf("inspect last_used_at: %v", err)
	}

	if lastUsed == initialLastUsed {
		t.Fatalf("last_used_at not advanced; still %q", lastUsed)
	}
}

func TestDeleteSessionAcceptsBlankAsNoOp(t *testing.T) {
	store := newTestStore(t)

	if err := store.Sessions.DeleteSession(context.Background(), ""); err != nil {
		t.Fatalf("blank delete should be a no-op, got %v", err)
	}

	if err := store.Sessions.DeleteSession(context.Background(), "   "); err != nil {
		t.Fatalf("whitespace delete should be a no-op, got %v", err)
	}
}

func TestDeleteSessionIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	session, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := store.Sessions.DeleteSession(ctx, session.RawToken); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	if err := store.Sessions.DeleteSession(ctx, session.RawToken); err != nil {
		t.Fatalf("second delete should still succeed, got %v", err)
	}
}

func TestHashTokenIsDeterministicAndHex(t *testing.T) {
	a := hashToken("foo")
	b := hashToken("foo")

	if a != b {
		t.Fatalf("hashToken not deterministic: %q vs %q", a, b)
	}

	if len(a) != 64 {
		t.Fatalf("expected 32-byte sha256 in hex (64 chars), got %d", len(a))
	}

	if strings.ContainsAny(a, "ghijklmnopqrstuvwxyz") {
		t.Fatalf("hex string must not contain non-hex chars: %q", a)
	}
}
