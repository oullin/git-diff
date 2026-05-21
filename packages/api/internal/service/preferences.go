package service

import (
	"context"

	"github.com/gocanto/git-diff/internal/storage"
)

// PreferenceService is a thin owner for the per-user UI preference map.
// The storage layer already handles the optimistic merge + delete-on-empty
// behaviour; the service exists so handlers can drop the store-handle
// boilerplate.
type PreferenceService struct {
	preferences *storage.PreferenceRepo
}

func NewPreferenceService(preferences *storage.PreferenceRepo) *PreferenceService {
	return &PreferenceService{preferences: preferences}
}

func (s *PreferenceService) Get(ctx context.Context, userID int64) (storage.UIPreferences, error) {
	if userID == 0 {
		return storage.UIPreferences{}, ErrAuthenticationRequired
	}

	return s.preferences.GetUIPreferences(ctx, userID)
}

func (s *PreferenceService) Save(
	ctx context.Context,
	userID int64,
	patch map[string]string,
) (storage.UIPreferences, error) {
	if userID == 0 {
		return storage.UIPreferences{}, ErrAuthenticationRequired
	}

	return s.preferences.SaveUIPreferences(ctx, userID, patch)
}
