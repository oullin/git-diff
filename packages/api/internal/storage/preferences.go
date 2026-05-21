package storage

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type UIPreferences struct {
	Values    map[string]string `json:"values"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

type PreferenceRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
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

func newPreferenceRepo(conn *sql.DB, queries *db.Queries, clk *clock) *PreferenceRepo {
	return &PreferenceRepo{db: conn, queries: queries, clk: clk}
}

func (r *PreferenceRepo) GetUIPreferences(ctx context.Context, userID int64) (UIPreferences, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT key, value, updated_at
		FROM ui_preferences
		WHERE user_id = ?
	`, userID)

	if err != nil {
		return UIPreferences{}, err
	}

	defer rows.Close()

	prefs := UIPreferences{Values: map[string]string{}}

	for rows.Next() {
		var (
			key       string
			value     string
			updatedAt string
		)

		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			return UIPreferences{}, err
		}

		prefs.Values[key] = value

		if updatedAt > prefs.UpdatedAt {
			prefs.UpdatedAt = updatedAt
		}
	}

	return prefs, rows.Err()
}

func (r *PreferenceRepo) SaveUIPreferences(ctx context.Context, userID int64, patch map[string]string) (UIPreferences, error) {
	if len(patch) == 0 {
		return r.GetUIPreferences(ctx, userID)
	}

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return UIPreferences{}, err
	}

	defer tx.Rollback()

	updatedAt := r.clk.now().UTC().Format(time.RFC3339Nano)

	for key, value := range patch {
		key = strings.TrimSpace(key)

		if key == "" {
			continue
		}

		if value == "" {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM ui_preferences WHERE user_id = ? AND key = ?
			`, userID, key); err != nil {
				return UIPreferences{}, err
			}

			continue
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO ui_preferences (user_id, key, value, updated_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, key) DO UPDATE SET
				value = excluded.value,
				updated_at = excluded.updated_at
		`, userID, key, value, updatedAt); err != nil {
			return UIPreferences{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return UIPreferences{}, err
	}

	return r.GetUIPreferences(ctx, userID)
}
