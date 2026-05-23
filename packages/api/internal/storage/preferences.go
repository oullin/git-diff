package storage

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UIPreferences struct {
	Values    map[string]string `json:"values"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

type PreferenceRepo struct {
	db  *gorm.DB
	clk *clock
}

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

func newPreferenceRepo(db *gorm.DB, clk *clock) *PreferenceRepo {
	return &PreferenceRepo{db: db, clk: clk}
}

func (r *PreferenceRepo) GetUIPreferences(ctx context.Context, userID int64) (UIPreferences, error) {
	var rows []UIPreferenceRow

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return UIPreferences{}, err
	}

	prefs := UIPreferences{Values: map[string]string{}}

	for _, row := range rows {
		prefs.Values[row.Key] = row.Value

		if row.UpdatedAt > prefs.UpdatedAt {
			prefs.UpdatedAt = row.UpdatedAt
		}
	}

	return prefs, nil
}

// SaveUIPreferences applies a patch in one transaction with at most two
// statements: a single batched upsert for set values and a single bulk delete
// for cleared values. Avoids the per-key INSERT/DELETE loop the original
// implementation issued.
func (r *PreferenceRepo) SaveUIPreferences(ctx context.Context, userID int64, patch map[string]string) (UIPreferences, error) {
	if len(patch) == 0 {
		return r.GetUIPreferences(ctx, userID)
	}

	updatedAt := r.clk.now().UTC().Format(time.RFC3339Nano)

	var (
		upserts []UIPreferenceRow
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

		upserts = append(upserts, UIPreferenceRow{
			UserID:    userID,
			Key:       key,
			Value:     value,
			UpdatedAt: updatedAt,
		})
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(upserts) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(&upserts).Error; err != nil {
				return err
			}
		}

		if len(deletes) > 0 {
			if err := tx.Where("user_id = ? AND key IN ?", userID, deletes).Delete(&UIPreferenceRow{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return UIPreferences{}, err
	}

	return r.GetUIPreferences(ctx, userID)
}
