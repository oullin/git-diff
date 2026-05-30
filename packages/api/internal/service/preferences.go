package service

import (
	"context"

	"github.com/oullin/git-diff/internal/storage"
)

type PreferenceService struct {
	preferences *storage.PreferenceRepo
}

func NewPreferenceService(preferences *storage.PreferenceRepo) *PreferenceService {
	return &PreferenceService{preferences: preferences}
}

func (s *PreferenceService) Get(ctx context.Context, userID int64) (storage.UserPreferences, error) {
	if userID == 0 {
		return storage.UserPreferences{}, ErrAuthenticationRequired
	}

	return s.preferences.GetUserPreferences(ctx, userID)
}

func (s *PreferenceService) Save(
	ctx context.Context,
	userID int64,
	patch map[string]string,
) (storage.UserPreferences, error) {
	if userID == 0 {
		return storage.UserPreferences{}, ErrAuthenticationRequired
	}

	return s.preferences.SaveUserPreferences(ctx, userID, patch)
}
