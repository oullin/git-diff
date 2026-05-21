package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrRepositoryNotFound = errors.New("repository not found")

func (s *Store) ListRepositoryCollaborators(ctx context.Context, ownerID int64, path string) ([]RepositoryCollaborator, error) {
	if err := s.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT ru.user_id, u.os_username, u.display_name, ru.role, ru.granted_at
		FROM repository_users ru
		JOIN users u ON u.id = ru.user_id
		WHERE ru.repo_path = ?
		ORDER BY u.os_username ASC
	`, path)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	collaborators := []RepositoryCollaborator{}

	for rows.Next() {
		var collaborator RepositoryCollaborator

		if err := rows.Scan(&collaborator.UserID, &collaborator.OSUsername, &collaborator.DisplayName, &collaborator.Role, &collaborator.GrantedAt); err != nil {
			return nil, err
		}

		collaborators = append(collaborators, collaborator)
	}

	return collaborators, rows.Err()
}

func (s *Store) GrantRepositoryAccess(ctx context.Context, ownerID int64, path string, userID int64, role string) (RepositoryCollaborator, error) {
	if userID == 0 {
		return RepositoryCollaborator{}, errors.New("user id is required")
	}

	if role != RepoRoleWrite && role != RepoRoleRead {
		return RepositoryCollaborator{}, fmt.Errorf("invalid role %q", role)
	}

	if err := s.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return RepositoryCollaborator{}, err
	}

	if userID == ownerID {
		return RepositoryCollaborator{}, errors.New("owner cannot also be a collaborator")
	}

	now := s.now().UTC().Format(time.RFC3339Nano)

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO repository_users (repo_path, user_id, role, granted_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(repo_path, user_id) DO UPDATE SET
			role = excluded.role,
			granted_at = excluded.granted_at
	`, path, userID, role, now); err != nil {
		return RepositoryCollaborator{}, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT ru.user_id, u.os_username, u.display_name, ru.role, ru.granted_at
		FROM repository_users ru
		JOIN users u ON u.id = ru.user_id
		WHERE ru.repo_path = ? AND ru.user_id = ?
	`, path, userID)

	var collaborator RepositoryCollaborator

	if err := row.Scan(&collaborator.UserID, &collaborator.OSUsername, &collaborator.DisplayName, &collaborator.Role, &collaborator.GrantedAt); err != nil {
		return RepositoryCollaborator{}, err
	}

	return collaborator, nil
}

func (s *Store) RevokeRepositoryAccess(ctx context.Context, ownerID int64, path string, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	if err := s.assertRepositoryOwner(ctx, ownerID, path); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, `
		DELETE FROM repository_users WHERE repo_path = ? AND user_id = ?
	`, path, userID)

	return err
}

func (s *Store) assertRepositoryOwner(ctx context.Context, ownerID int64, path string) error {
	if ownerID == 0 {
		return errors.New("user id is required")
	}

	if path == "" {
		return errors.New("repository path is required")
	}

	var actualOwner int64
	err := s.db.QueryRowContext(ctx, `SELECT owner_id FROM repositories WHERE path = ?`, path).Scan(&actualOwner)

	if errors.Is(err, sql.ErrNoRows) {
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
