package storage

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Branch struct {
	Name       string `json:"name"`
	Locked     bool   `json:"locked"`
	LockedBy   int64  `json:"lockedBy,omitempty"`
	LockedAt   string `json:"lockedAt,omitempty"`
	LastSeenAt string `json:"lastSeenAt"`
}

type BranchRepo struct {
	db  *gorm.DB
	clk *clock
}

func newBranchRepo(db *gorm.DB, clk *clock) *BranchRepo {
	return &BranchRepo{db: db, clk: clk}
}

// SyncBranches reconciles the branches table for a repo to exactly `names` in
// one transaction. Issues two statements at most: one batched upsert and one
// delete of stale rows. Previously did a per-branch INSERT in a loop.
func (r *BranchRepo) SyncBranches(ctx context.Context, path string, names []string) error {
	if path == "" {
		return errors.New("repository path is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	rows := make([]BranchRow, 0, len(names))

	for _, name := range names {
		if name == "" {
			continue
		}

		rows = append(rows, BranchRow{
			RepoPath:   path,
			Name:       name,
			LastSeenAt: now,
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(rows) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "repo_path"}, {Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"last_seen_at"}),
			}).Create(&rows).Error; err != nil {
				return err
			}
		}

		return tx.Where("repo_path = ? AND last_seen_at <> ? AND locked = 0", path, now).
			Delete(&BranchRow{}).Error
	})
}

func (r *BranchRepo) ListBranches(ctx context.Context, path string) ([]Branch, error) {
	if path == "" {
		return nil, errors.New("repository path is required")
	}

	var rows []BranchRow

	if err := r.db.WithContext(ctx).
		Where("repo_path = ?", path).
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

func (r *BranchRepo) LockBranch(ctx context.Context, path, name string, userID int64) error {
	if path == "" || name == "" {
		return errors.New("repository path and branch name are required")
	}

	if userID == 0 {
		return errors.New("user id is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	result := r.db.WithContext(ctx).
		Model(&BranchRow{}).
		Where("repo_path = ? AND name = ?", path, name).
		Updates(map[string]any{
			"locked":    1,
			"locked_by": userID,
			"locked_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *BranchRepo) UnlockBranch(ctx context.Context, path, name string) error {
	if path == "" || name == "" {
		return errors.New("repository path and branch name are required")
	}

	result := r.db.WithContext(ctx).
		Model(&BranchRow{}).
		Where("repo_path = ? AND name = ?", path, name).
		Updates(map[string]any{
			"locked":    0,
			"locked_by": nil,
			"locked_at": nil,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *BranchRepo) IsBranchLocked(ctx context.Context, path, name string) (bool, error) {
	var locked int

	err := r.db.WithContext(ctx).
		Model(&BranchRow{}).
		Select("locked").
		Where("repo_path = ? AND name = ?", path, name).
		Take(&locked).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return locked != 0, nil
}

func (r *BranchRepo) DeleteBranchRow(ctx context.Context, path, name string) error {
	return r.db.WithContext(ctx).
		Where("repo_path = ? AND name = ?", path, name).
		Delete(&BranchRow{}).Error
}
