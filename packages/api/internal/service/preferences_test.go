package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oullin/git-diff/internal/storage"
)

func newPreferenceService(t *testing.T, store *storage.Store) *PreferenceService {
	t.Helper()

	return NewPreferenceService(store.Preferences)
}

func TestPreferenceServiceRequiresAuth(t *testing.T) {
	store := newTestStore(t)
	svc := newPreferenceService(t, store)

	if _, err := svc.Get(context.Background(), 0); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Get should require auth, got %v", err)
	}

	if _, err := svc.Save(context.Background(), 0, map[string]string{"k": "v"}); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("Save should require auth, got %v", err)
	}
}

func TestPreferenceServiceSaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")
	svc := newPreferenceService(t, store)

	if _, err := svc.Save(ctx, user.ID, map[string]string{storage.PrefKeyTheme: "dark"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	prefs, err := svc.Get(ctx, user.ID)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if prefs.Values[storage.PrefKeyTheme] != "dark" {
		t.Fatalf("expected theme=dark, got %#v", prefs.Values)
	}
}
