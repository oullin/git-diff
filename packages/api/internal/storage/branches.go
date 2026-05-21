package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type Branch struct {
	Name       string `json:"name"`
	Locked     bool   `json:"locked"`
	LockedBy   int64  `json:"lockedBy,omitempty"`
	LockedAt   string `json:"lockedAt,omitempty"`
	LastSeenAt string `json:"lastSeenAt"`
}

type BranchRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
}

func newBranchRepo(conn *sql.DB, queries *db.Queries, clk *clock) *BranchRepo {
	return &BranchRepo{db: conn, queries: queries, clk: clk}
}

func (r *BranchRepo) SyncBranches(ctx context.Context, path string, names []string) error {
	if path == "" {
		return errors.New("repository path is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	for _, name := range names {
		if name == "" {
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO branches (repo_path, name, locked, last_seen_at)
			VALUES (?, ?, 0, ?)
			ON CONFLICT(repo_path, name) DO UPDATE SET
				last_seen_at = excluded.last_seen_at
		`, path, name, now); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM branches WHERE repo_path = ? AND last_seen_at <> ? AND locked = 0
	`, path, now); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BranchRepo) ListBranches(ctx context.Context, path string) ([]Branch, error) {
	if path == "" {
		return nil, errors.New("repository path is required")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT name, locked, locked_by, locked_at, last_seen_at
		FROM branches
		WHERE repo_path = ?
		ORDER BY name ASC
	`, path)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	branches := []Branch{}

	for rows.Next() {
		var (
			branch   Branch
			locked   int
			lockedBy sql.NullInt64
			lockedAt sql.NullString
		)

		if err := rows.Scan(&branch.Name, &locked, &lockedBy, &lockedAt, &branch.LastSeenAt); err != nil {
			return nil, err
		}

		branch.Locked = locked != 0

		if lockedBy.Valid {
			branch.LockedBy = lockedBy.Int64
		}

		branch.LockedAt = fromNull(lockedAt)
		branches = append(branches, branch)
	}

	return branches, rows.Err()
}

func (r *BranchRepo) LockBranch(ctx context.Context, path, name string, userID int64) error {
	if path == "" || name == "" {
		return errors.New("repository path and branch name are required")
	}

	if userID == 0 {
		return errors.New("user id is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	result, err := r.db.ExecContext(ctx, `
		UPDATE branches
		SET locked = 1, locked_by = ?, locked_at = ?
		WHERE repo_path = ? AND name = ?
	`, userID, now, path, name)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *BranchRepo) UnlockBranch(ctx context.Context, path, name string) error {
	if path == "" || name == "" {
		return errors.New("repository path and branch name are required")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE branches
		SET locked = 0, locked_by = NULL, locked_at = NULL
		WHERE repo_path = ? AND name = ?
	`, path, name)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("branch not found")
	}

	return nil
}

func (r *BranchRepo) IsBranchLocked(ctx context.Context, path, name string) (bool, error) {
	var locked int
	err := r.db.QueryRowContext(ctx, `
		SELECT locked FROM branches WHERE repo_path = ? AND name = ?
	`, path, name).Scan(&locked)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return locked != 0, nil
}

func (r *BranchRepo) DeleteBranchRow(ctx context.Context, path, name string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM branches WHERE repo_path = ? AND name = ?
	`, path, name)

	return err
}
