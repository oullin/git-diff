package storage

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gocanto/git-diff/internal/db"
)

type UserPreferences struct {
	Values    map[string]string `json:"values"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

type PreferenceRepo struct{}

const DefaultTheme = "light"
const DefaultDiffViewMode = "split"

const (
	PrefKeyTheme              = "theme"
	PrefKeyDiffViewMode       = "diff.viewMode"
	PrefKeyDiffHideWhitespace = "diff.hideWhitespace"
	PrefKeyLastRepoRoot       = "repo.lastRoot"
	PrefKeyPanelLeftWidth     = "panel.left.width"
	PrefKeyPanelFileTreeWidth = "panel.fileTree.width"
	PrefKeyPanelRightWidth    = "panel.right.width"
	PrefKeyAnthropicAPIKey    = "llm.anthropicApiKey"
	PrefKeyAnthropicModel     = "llm.anthropicModel"
)

func newPreferenceRepo() *PreferenceRepo {
	return &PreferenceRepo{}
}

func (r *PreferenceRepo) GetUserPreferences(ctx context.Context, userID int64) (UserPreferences, error) {
	var rows []UserPreferenceRow

	if err := db.Conn().WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return UserPreferences{}, err
	}

	prefs := UserPreferences{Values: map[string]string{}}

	for _, row := range rows {
		prefs.Values[row.Key] = row.Value

		if row.UpdatedAt > prefs.UpdatedAt {
			prefs.UpdatedAt = row.UpdatedAt
		}
	}

	return prefs, nil
}

// SaveUserPreferences applies a patch in one transaction with at most two
// statements: a single batched upsert for set values and a single bulk delete
// for cleared values. Avoids the per-key INSERT/DELETE loop the original
// implementation issued.
func (r *PreferenceRepo) SaveUserPreferences(ctx context.Context, userID int64, patch map[string]string) (UserPreferences, error) {
	if len(patch) == 0 {
		return r.GetUserPreferences(ctx, userID)
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	var (
		upserts []UserPreferenceRow
		deletes []string
	)

	for key, value := range patch {
		key = strings.TrimSpace(key)

		if key == "" {
			continue
		}

		if value == "" {
			deletes = append(deletes, key)

			continue
		}

		upserts = append(upserts, UserPreferenceRow{
			UserID:    userID,
			Key:       key,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	err := db.Conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(upserts) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(&upserts).Error; err != nil {
				return err
			}
		}

		if len(deletes) > 0 {
			if err := tx.Where("user_id = ? AND key IN ?", userID, deletes).Delete(&UserPreferenceRow{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return UserPreferences{}, err
	}

	return r.GetUserPreferences(ctx, userID)
}
