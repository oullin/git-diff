package storage

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gocanto/git-diff/internal/db"
)

type Branch struct {
	Name       string `json:"name"`
	Locked     bool   `json:"locked"`
	LockedBy   int64  `json:"lockedBy,omitempty"`
	LockedAt   string `json:"lockedAt,omitempty"`
	LastSeenAt string `json:"lastSeenAt"`
}

type RepositoryBranchRepo struct{}

func newRepositoryBranchRepo() *RepositoryBranchRepo {
	return &RepositoryBranchRepo{}
}

// SyncBranches reconciles the repository_branches table for a repo to exactly
// `names` in one transaction. Issues two statements at most: one batched upsert
// and one delete of stale rows.
func (r *RepositoryBranchRepo) SyncBranches(ctx context.Context, repositoryID int64, names []string) error {
	if repositoryID == 0 {
		return errors.New("repository id is required")
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	rows := make([]RepositoryBranchRow, 0, len(names))

	for _, name := range names {
		if name == "" {
			continue
		}

		rows = append(rows, RepositoryBranchRow{
			RepositoryID: repositoryID,
			Name:         name,
			LastSeenAt:   now,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	return db.Conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(rows) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "repository_id"}, {Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"last_seen_at", "updated_at"}),
			}).Create(&rows).Error; err != nil {
				return err
			}
		}

		return tx.Where("repository_id = ? AND last_seen_at <> ? AND locked = 0", repositoryID, now).
			Delete(&RepositoryBranchRow{}).Error
	})
}

func (r *RepositoryBranchRepo) ListBranches(ctx context.Context, repositoryID int64) ([]Branch, error) {
	if repositoryID == 0 {
		return nil, errors.New("repository id is required")
	}

	var rows []RepositoryBranchRow

	if err := db.Conn().WithContext(ctx).
		Where("repository_id = ?", repositoryID).
		Order("name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	branches := make([]Branch, 0, len(rows))

	for _, row := range rows {
		branch := Branch{
			Name:       row.Name,
			Locked:     row.Locked != 0,
			LastSeenAt: row.LastSeenAt,
		}

		if row.LockedBy != nil {
			branch.LockedBy = *row.LockedBy
		}

		if row.LockedAt != nil {
			branch.LockedAt = *row.LockedAt
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func (r *RepositoryBranchRepo) LockBranch(ctx context.Context, repositoryID int64, name string, userID int64) error {
	if repositoryID == 0 || name == "" {
		return errors.New("repository id and branch name are required")
	}

	if userID == 0 {
		return errors.New("user id is required")
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	result := db.Conn().WithContext(ctx).
		Model(&RepositoryBranchRow{}).
		Where("repository_id = ? AND name = ?", repositoryID, name).
		Updates(map[string]any{
			"locked":     1,
			"locked_by":  userID,
			"locked_at":  now,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *RepositoryBranchRepo) UnlockBranch(ctx context.Context, repositoryID int64, name string) error {
	if repositoryID == 0 || name == "" {
		return errors.New("repository id and branch name are required")
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)

	result := db.Conn().WithContext(ctx).
		Model(&RepositoryBranchRow{}).
		Where("repository_id = ? AND name = ?", repositoryID, name).
		Updates(map[string]any{
			"locked":     0,
			"locked_by":  nil,
			"locked_at":  nil,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *RepositoryBranchRepo) IsBranchLocked(ctx context.Context, repositoryID int64, name string) (bool, error) {
	var locked int

	err := db.Conn().WithContext(ctx).
		Model(&RepositoryBranchRow{}).
		Select("locked").
		Where("repository_id = ? AND name = ?", repositoryID, name).
		Take(&locked).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return locked != 0, nil
}

func (r *RepositoryBranchRepo) DeleteBranchRow(ctx context.Context, repositoryID int64, name string) error {
	return db.Conn().WithContext(ctx).
		Where("repository_id = ? AND name = ?", repositoryID, name).
		Delete(&RepositoryBranchRow{}).Error
}
