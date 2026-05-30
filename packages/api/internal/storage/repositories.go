package storage

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oullin/git-diff/internal/db"
)

type Repository struct {
	ID           int64  `json:"id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	OwnerID      int64  `json:"ownerId"`
	Role         string `json:"role"`
	AddedAt      string `json:"addedAt"`
	LastOpenedAt string `json:"lastOpenedAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type RepoRepo struct{}

// repoListRow holds the columns returned by the owner-or-collaborator LEFT
// JOIN below; the join is written as one query to keep listing O(1) statements.
type repoListRow struct {
	ID           int64
	Path         string
	Name         string
	OwnerID      int64
	AddedAt      string
	LastOpenedAt *string
	CreatedAt    string
	UpdatedAt    string
	Role         *string
}

const (
	RepoRoleOwner = "owner"
	RepoRoleWrite = "write"
	RepoRoleRead  = "read"
)

var ErrRepositoryNotOwned = errors.New("repository not found or not owned by user")

func newRepoRepo() *RepoRepo {
	return &RepoRepo{}
}

func (r *RepoRepo) ListRepositoriesForUser(ctx context.Context, userID int64) ([]Repository, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	var rows []repoListRow

	err := db.Conn().WithContext(ctx).
		Table("repositories AS r").
		Select(`r.id, r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			r.created_at, r.updated_at,
			CASE WHEN r.owner_id = ? THEN 'owner' ELSE ru.role END AS role`, userID).
		Joins("LEFT JOIN repository_users ru ON ru.repository_id = r.id AND ru.user_id = ?", userID).
		Where("r.owner_id = ? OR ru.user_id = ?", userID, userID).
		Order("COALESCE(r.last_opened_at, r.added_at) DESC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	repos := make([]Repository, 0, len(rows))

	for _, row := range rows {
		repos = append(repos, toRepository(row))
	}

	return repos, nil
}

func (r *RepoRepo) GetRepository(ctx context.Context, userID int64, path string) (Repository, error) {
	if userID == 0 {
		return Repository{}, errors.New("user id is required")
	}

	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	var row repoListRow

	err := db.Conn().WithContext(ctx).
		Table("repositories AS r").
		Select(`r.id, r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			r.created_at, r.updated_at,
			CASE WHEN r.owner_id = ? THEN 'owner' ELSE ru.role END AS role`, userID).
		Joins("LEFT JOIN repository_users ru ON ru.repository_id = r.id AND ru.user_id = ?", userID).
		Where("r.path = ? AND (r.owner_id = ? OR ru.user_id = ?)", path, userID, userID).
		Scan(&row).Error

	if err != nil {
		return Repository{}, err
	}

	if row.Path == "" {
		return Repository{}, gorm.ErrRecordNotFound
	}

	return toRepository(row), nil
}

// GetByPath returns the bare row for a path, without joining the role of the
// requesting user. Used by collaborator wiring where we already authorized the
// caller as the owner.
func (r *RepoRepo) GetByPath(ctx context.Context, path string) (Repository, error) {
	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	var row RepositoryRow

	if err := db.Conn().WithContext(ctx).Where("path = ?", path).Take(&row).Error; err != nil {
		return Repository{}, err
	}

	repo := Repository{
		ID:        row.ID,
		Path:      row.Path,
		Name:      row.Name,
		OwnerID:   row.OwnerID,
		AddedAt:   row.AddedAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	if row.LastOpenedAt != nil {
		repo.LastOpenedAt = *row.LastOpenedAt
	}

	return repo, nil
}

func (r *RepoRepo) UpsertRepository(ctx context.Context, ownerID int64, path string, name string) (Repository, error) {
	if ownerID == 0 {
		return Repository{}, errors.New("owner id is required")
	}

	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	if name == "" {
		name = filepath.Base(path)
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)
	row := RepositoryRow{
		Path:         path,
		Name:         name,
		OwnerID:      ownerID,
		AddedAt:      now,
		LastOpenedAt: &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := db.Conn().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "path"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "last_opened_at", "updated_at"}),
		}).
		Create(&row).Error

	if err != nil {
		return Repository{}, err
	}

	return r.GetRepository(ctx, ownerID, path)
}

func (r *RepoRepo) RemoveRepository(ctx context.Context, userID int64, path string) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	if path == "" {
		return errors.New("repository path is required")
	}

	result := db.Conn().WithContext(ctx).
		Where("path = ? AND owner_id = ?", path, userID).
		Delete(&RepositoryRow{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRepositoryNotOwned
	}

	return nil
}

func toRepository(row repoListRow) Repository {
	repo := Repository{
		ID:        row.ID,
		Path:      row.Path,
		Name:      row.Name,
		OwnerID:   row.OwnerID,
		AddedAt:   row.AddedAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	if row.LastOpenedAt != nil {
		repo.LastOpenedAt = *row.LastOpenedAt
	}

	if row.Role != nil {
		repo.Role = *row.Role
	}

	return repo
}
