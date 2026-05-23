package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepositoryCollaborator struct {
	UserID      int64  `json:"userId"`
	OSUsername  string `json:"osUsername"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	GrantedAt   string `json:"grantedAt"`
}

type CollaboratorRepo struct {
	db  *gorm.DB
	clk *clock
}

var ErrRepositoryNotFound = errors.New("repository not found")

func newCollaboratorRepo(db *gorm.DB, clk *clock) *CollaboratorRepo {
	return &CollaboratorRepo{db: db, clk: clk}
}

func (r *CollaboratorRepo) List(ctx context.Context, ownerID int64, path string) ([]RepositoryCollaborator, error) {
	if err := r.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return nil, err
	}

	var rows []RepositoryCollaborator

	err := r.db.WithContext(ctx).Raw(`
		SELECT ru.user_id AS user_id, u.os_username AS os_username, u.display_name AS display_name, ru.role AS role, ru.granted_at AS granted_at
		FROM repository_users ru
		JOIN users u ON u.id = ru.user_id
		WHERE ru.repo_path = ?
		ORDER BY u.os_username ASC
	`, path).Scan(&rows).Error

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

	if err := r.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return RepositoryCollaborator{}, err
	}

	if userID == ownerID {
		return RepositoryCollaborator{}, errors.New("owner cannot also be a collaborator")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	row := RepositoryUserRow{
		RepoPath:  path,
		UserID:    userID,
		Role:      role,
		GrantedAt: now,
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "repo_path"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role", "granted_at"}),
		}).
		Create(&row).Error; err != nil {
		return RepositoryCollaborator{}, err
	}

	var collaborator RepositoryCollaborator

	err := r.db.WithContext(ctx).Raw(`
		SELECT ru.user_id AS user_id, u.os_username AS os_username, u.display_name AS display_name, ru.role AS role, ru.granted_at AS granted_at
		FROM repository_users ru
		JOIN users u ON u.id = ru.user_id
		WHERE ru.repo_path = ? AND ru.user_id = ?
	`, path, userID).Scan(&collaborator).Error

	if err != nil {
		return RepositoryCollaborator{}, err
	}

	return collaborator, nil
}

func (r *CollaboratorRepo) Revoke(ctx context.Context, ownerID int64, path string, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	if err := r.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Where("repo_path = ? AND user_id = ?", path, userID).
		Delete(&RepositoryUserRow{}).Error
}

func (r *CollaboratorRepo) assertRepositoryOwner(ctx context.Context, ownerID int64, path string) error {
	if ownerID == 0 {
		return errors.New("user id is required")
	}

	if path == "" {
		return errors.New("repository path is required")
	}

	var actualOwner int64

	err := r.db.WithContext(ctx).
		Model(&RepositoryRow{}).
		Select("owner_id").
		Where("path = ?", path).
		Take(&actualOwner).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRepositoryNotFound
	}

	if err != nil {
		return err
	}

	if actualOwner != ownerID {
		return ErrRepositoryNotOwned
	}

	return nil
}
