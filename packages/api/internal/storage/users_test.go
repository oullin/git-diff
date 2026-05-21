package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.sqlite3"))

	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	return store
}

func TestEnsureUserIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	first, err := store.Users.EnsureUser(ctx, "alice")

	if err != nil {
		t.Fatalf("first ensure: %v", err)
	}

	second, err := store.Users.EnsureUser(ctx, "alice")

	if err != nil {
		t.Fatalf("second ensure: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same id, got %d vs %d", first.ID, second.ID)
	}

	if first.HasPassword || second.HasPassword {
		t.Fatalf("expected fresh user to have no password")
	}
}

func TestSaveUIPreferencesUpsertsAndDeletesOnEmpty(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	user, err := store.Users.EnsureUser(ctx, "bob")

	if err != nil {
		t.Fatalf("ensure: %v", err)
	}

	if _, err := store.Preferences.SaveUIPreferences(ctx, user.ID, map[string]string{
		PrefKeyTheme:          "dark",
		PrefKeyPanelLeftWidth: "30",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	prefs, err := store.Preferences.GetUIPreferences(ctx, user.ID)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if prefs.Values[PrefKeyTheme] != "dark" || prefs.Values[PrefKeyPanelLeftWidth] != "30" {
		t.Fatalf("unexpected values after upsert: %#v", prefs.Values)
	}

	if _, err := store.Preferences.SaveUIPreferences(ctx, user.ID, map[string]string{
		PrefKeyTheme: "",
	}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	prefs, err = store.Preferences.GetUIPreferences(ctx, user.ID)

	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}

	if _, present := prefs.Values[PrefKeyTheme]; present {
		t.Fatalf("expected theme key removed, got %#v", prefs.Values)
	}

	if prefs.Values[PrefKeyPanelLeftWidth] != "30" {
		t.Fatalf("expected panel width retained, got %#v", prefs.Values)
	}
}

func TestSessionCreateResumeExpireAndDelete(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	user, err := store.Users.EnsureUser(ctx, "carol")

	if err != nil {
		t.Fatalf("ensure: %v", err)
	}

	session, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if session.RawToken == "" {
		t.Fatalf("expected raw token")
	}

	resumed, err := store.Sessions.ResumeSession(ctx, store.Users, session.RawToken)

	if err != nil {
		t.Fatalf("resume: %v", err)
	}

	if resumed.ID != user.ID {
		t.Fatalf("resumed user mismatch: %d vs %d", resumed.ID, user.ID)
	}

	store.SetNow(func() time.Time { return time.Now().Add(2 * time.Hour) })

	if _, err := store.Sessions.ResumeSession(ctx, store.Users, session.RawToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected expired session to be rejected, got %v", err)
	}

	store.SetNow(time.Now)

	fresh, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create fresh session: %v", err)
	}

	if err := store.Sessions.DeleteSession(ctx, fresh.RawToken); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	if _, err := store.Sessions.ResumeSession(ctx, store.Users, fresh.RawToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected deleted session to be rejected, got %v", err)
	}
}

func TestWipeUserCascadesPreferencesAndSessions(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	user, err := store.Users.EnsureUser(ctx, "dave")

	if err != nil {
		t.Fatalf("ensure: %v", err)
	}

	if _, err := store.Preferences.SaveUIPreferences(ctx, user.ID, map[string]string{
		PrefKeyTheme: "dark",
	}); err != nil {
		t.Fatalf("save prefs: %v", err)
	}

	session, err := store.Sessions.CreateSession(ctx, user.ID, time.Hour)

	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := store.Users.WipeUser(ctx, user.ID); err != nil {
		t.Fatalf("wipe user: %v", err)
	}

	if _, err := store.Sessions.ResumeSession(ctx, store.Users, session.RawToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected cascading session delete, got %v", err)
	}

	prefs, err := store.Preferences.GetUIPreferences(ctx, user.ID)

	if err != nil {
		t.Fatalf("get prefs after wipe: %v", err)
	}

	if len(prefs.Values) != 0 {
		t.Fatalf("expected empty prefs after cascade, got %#v", prefs.Values)
	}
}

func TestSetPasswordHashMarksUserAsHavingPassword(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	user, err := store.Users.EnsureUser(ctx, "eve")

	if err != nil {
		t.Fatalf("ensure: %v", err)
	}

	if err := store.Users.SetPasswordHash(ctx, user.ID, "$2a$12$exampleexampleexampleexampleexampleexampleexampleexample"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	updated, err := store.Users.GetUserByID(ctx, user.ID)

	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if !updated.HasPassword {
		t.Fatalf("expected HasPassword=true, got %#v", updated)
	}
}
