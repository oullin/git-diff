package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type Repository struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	OwnerID      int64  `json:"ownerId"`
	Role         string `json:"role"`
	AddedAt      string `json:"addedAt"`
	LastOpenedAt string `json:"lastOpenedAt,omitempty"`
}

type RepoRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

const (
	RepoRoleOwner = "owner"
	RepoRoleWrite = "write"
	RepoRoleRead  = "read"
)

var ErrRepositoryNotOwned = errors.New("repository not found or not owned by user")

func newRepoRepo(conn *sql.DB, queries *db.Queries, clk *clock) *RepoRepo {
	return &RepoRepo{db: conn, queries: queries, clk: clk}
}

func (r *RepoRepo) ListRepositoriesForUser(ctx context.Context, userID int64) ([]Repository, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			CASE WHEN r.owner_id = ?1 THEN 'owner' ELSE ru.role END AS role
		FROM repositories r
		LEFT JOIN repository_users ru ON ru.repo_path = r.path AND ru.user_id = ?1
		WHERE r.owner_id = ?1 OR ru.user_id = ?1
		ORDER BY COALESCE(r.last_opened_at, r.added_at) DESC
	`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := []Repository{}

	for rows.Next() {
		var (
			repo         Repository
			lastOpenedAt sql.NullString
			role         sql.NullString
		)

		if err := rows.Scan(&repo.Path, &repo.Name, &repo.OwnerID, &repo.AddedAt, &lastOpenedAt, &role); err != nil {
			return nil, err
		}

		repo.LastOpenedAt = fromNull(lastOpenedAt)
		repo.Role = fromNull(role)
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

func (r *RepoRepo) GetRepository(ctx context.Context, userID int64, path string) (Repository, error) {
	if userID == 0 {
		return Repository{}, errors.New("user id is required")
	}

	if path == "" {
		return Repository{}, errors.New("repository path is required")
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT r.path, r.name, r.owner_id, r.added_at, r.last_opened_at,
			CASE WHEN r.owner_id = ?1 THEN 'owner' ELSE ru.role END AS role
		FROM repositories r
		LEFT JOIN repository_users ru ON ru.repo_path = r.path AND ru.user_id = ?1
		WHERE r.path = ?2 AND (r.owner_id = ?1 OR ru.user_id = ?1)
	`, userID, path)

	var (
		repo         Repository
		lastOpenedAt sql.NullString
		role         sql.NullString
	)

	if err := row.Scan(&repo.Path, &repo.Name, &repo.OwnerID, &repo.AddedAt, &lastOpenedAt, &role); err != nil {
		return Repository{}, err
	}

	repo.LastOpenedAt = fromNull(lastOpenedAt)
	repo.Role = fromNull(role)

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

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO repositories (path, name, owner_id, added_at, last_opened_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			last_opened_at = excluded.last_opened_at
	`, path, name, ownerID, now, now)

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

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM repositories WHERE path = ? AND owner_id = ?
	`, path, userID)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrRepositoryNotOwned
	}

	return nil
}
