package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oullin/git-diff/internal/db"
)

type RepositoryCollaborator struct {
	UserID      int64  `json:"userId"`
	OSUsername  string `json:"osUsername"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	GrantedAt   string `json:"grantedAt"`
}

type CollaboratorRepo struct{}

var ErrRepositoryNotFound = errors.New("repository not found")

func newCollaboratorRepo() *CollaboratorRepo {
	return &CollaboratorRepo{}
}

func (r *CollaboratorRepo) List(ctx context.Context, ownerID int64, path string) ([]RepositoryCollaborator, error) {
	repoID, err := r.assertRepositoryOwner(ctx, ownerID, path)

	if err != nil {
		return nil, err
	}

	var rows []RepositoryCollaborator

	err = db.Conn().WithContext(ctx).
		Table("repository_users AS ru").
		Select("ru.user_id AS user_id, u.os_username AS os_username, u.display_name AS display_name, ru.role AS role, ru.granted_at AS granted_at").
		Joins("JOIN users u ON u.id = ru.user_id").
		Where("ru.repository_id = ?", repoID).
		Order("u.os_username ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	if rows == nil {
		rows = []RepositoryCollaborator{}
	}

	return rows, nil
}

func (r *CollaboratorRepo) Grant(ctx context.Context, ownerID int64, path string, userID int64, role string) (RepositoryCollaborator, error) {
	if userID == 0 {
		return RepositoryCollaborator{}, errors.New("user id is required")
	}

	if role != RepoRoleWrite && role != RepoRoleRead {
		return RepositoryCollaborator{}, fmt.Errorf("invalid role %q", role)
	}

	repoID, err := r.assertRepositoryOwner(ctx, ownerID, path)

	if err != nil {
		return RepositoryCollaborator{}, err
	}

	if userID == ownerID {
		return RepositoryCollaborator{}, errors.New("owner cannot also be a collaborator")
	}

	now := db.Now().UTC().Format(time.RFC3339Nano)
	row := RepositoryUserRow{
		RepositoryID: repoID,
		UserID:       userID,
		Role:         role,
		GrantedAt:    now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := db.Conn().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "repository_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role", "granted_at", "updated_at"}),
		}).
		Create(&row).Error; err != nil {
		return RepositoryCollaborator{}, err
	}

	var collaborator RepositoryCollaborator

	err = db.Conn().WithContext(ctx).
		Table("repository_users AS ru").
		Select("ru.user_id AS user_id, u.os_username AS os_username, u.display_name AS display_name, ru.role AS role, ru.granted_at AS granted_at").
		Joins("JOIN users u ON u.id = ru.user_id").
		Where("ru.repository_id = ? AND ru.user_id = ?", repoID, userID).
		Scan(&collaborator).Error

	if err != nil {
		return RepositoryCollaborator{}, err
	}

	return collaborator, nil
}

func (r *CollaboratorRepo) Revoke(ctx context.Context, ownerID int64, path string, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	repoID, err := r.assertRepositoryOwner(ctx, ownerID, path)

	if err != nil {
		return err
	}

	return db.Conn().WithContext(ctx).
		Where("repository_id = ? AND user_id = ?", repoID, userID).
		Delete(&RepositoryUserRow{}).Error
}

// assertRepositoryOwner returns the numeric repository_id once it has confirmed
// ownerID owns the row at path. Returns ErrRepositoryNotFound when no row
// exists and ErrRepositoryNotOwned when the row belongs to someone else.
func (r *CollaboratorRepo) assertRepositoryOwner(ctx context.Context, ownerID int64, path string) (int64, error) {
	if ownerID == 0 {
		return 0, errors.New("user id is required")
	}

	if path == "" {
		return 0, errors.New("repository path is required")
	}

	var row struct {
		ID      int64
		OwnerID int64
	}

	err := db.Conn().WithContext(ctx).
		Model(&RepositoryRow{}).
		Select("id", "owner_id").
		Where("path = ?", path).
		Take(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrRepositoryNotFound
	}

	if err != nil {
		return 0, err
	}

	if row.OwnerID != ownerID {
		return 0, ErrRepositoryNotOwned
	}

	return row.ID, nil
}
