package storage

import (
	"context"
	"testing"
)

func TestSaveUserPreferencesUpsertsAndDeletesInOneCall(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	if _, err := store.Preferences.SaveUserPreferences(ctx, user.ID, map[string]string{
		PrefKeyTheme:        "dark",
		PrefKeyDiffViewMode: "unified",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	prefs, err := store.Preferences.GetUserPreferences(ctx, user.ID)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if prefs.Values[PrefKeyTheme] != "dark" || prefs.Values[PrefKeyDiffViewMode] != "unified" {
		t.Fatalf("upsert lost data: %#v", prefs.Values)
	}

	// Mixed patch: clear theme, set new key.
	if _, err := store.Preferences.SaveUserPreferences(ctx, user.ID, map[string]string{
		PrefKeyTheme:          "",
		PrefKeyPanelLeftWidth: "30",
	}); err != nil {
		t.Fatalf("patch: %v", err)
	}

	prefs, _ = store.Preferences.GetUserPreferences(ctx, user.ID)

	if _, present := prefs.Values[PrefKeyTheme]; present {
		t.Fatalf("expected theme cleared, got %#v", prefs.Values)
	}

	if prefs.Values[PrefKeyPanelLeftWidth] != "30" {
		t.Fatalf("expected panel.left.width=30, got %#v", prefs.Values)
	}

	if prefs.Values[PrefKeyDiffViewMode] != "unified" {
		t.Fatalf("expected previous value retained, got %#v", prefs.Values)
	}
}

func TestSaveUserPreferencesEmptyPatchIsNoOp(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	if _, err := store.Preferences.SaveUserPreferences(ctx, user.ID, map[string]string{PrefKeyTheme: "dark"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	prefs, err := store.Preferences.SaveUserPreferences(ctx, user.ID, map[string]string{})

	if err != nil {
		t.Fatalf("empty patch: %v", err)
	}

	if prefs.Values[PrefKeyTheme] != "dark" {
		t.Fatalf("empty patch must not modify state, got %#v", prefs.Values)
	}
}

func TestSaveUserPreferencesTrimsAndIgnoresBlankKeys(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	if _, err := store.Preferences.SaveUserPreferences(ctx, user.ID, map[string]string{
		"   ":           "ignored",
		"":              "ignored",
		" theme   ":     "dark",
		"diff.viewMode": "split",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	prefs, _ := store.Preferences.GetUserPreferences(ctx, user.ID)

	if prefs.Values["theme"] != "dark" {
		t.Fatalf("expected trimmed key 'theme' to be set, got %#v", prefs.Values)
	}

	for k := range prefs.Values {
		if k == "" {
			t.Fatalf("blank key persisted: %#v", prefs.Values)
		}
	}
}

func TestGetUserPreferencesEmpty(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := seedUser(t, store, "alice")

	prefs, err := store.Preferences.GetUserPreferences(ctx, user.ID)

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if prefs.Values == nil {
		t.Fatalf("expected non-nil map")
	}

	if len(prefs.Values) != 0 {
		t.Fatalf("expected empty map, got %#v", prefs.Values)
	}
}
