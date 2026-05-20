package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// PendingComment is a draft comment that exists before a review session has
// been created. The (user_id, repo_root, context_kind, context_sha) tuple
// scopes it; when the user starts a review, PromotePendingComments moves any
// matching rows into review_comments.
type PendingComment struct {
	ID          string `json:"id"`
	UserID      int64  `json:"userId"`
	RepoRoot    string `json:"repoRoot"`
	ContextKind string `json:"contextKind"`
	ContextSHA  string `json:"contextSha,omitempty"`
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type PendingCommentInput struct {
	RepoRoot    string `json:"repoRoot"`
	ContextKind string `json:"contextKind"`
	ContextSHA  string `json:"contextSha"`
	FilePath    string `json:"filePath"`
	DiffSection string `json:"diffSection"`
	Side        string `json:"side"`
	LineNumber  int64  `json:"lineNumber"`
	AuthorLabel string `json:"authorLabel"`
	BodyHTML    string `json:"bodyHtml"`
}

func (s *Store) CreatePendingComment(ctx context.Context, userID int64, id string, input PendingCommentInput) (PendingComment, error) {
	if userID == 0 {
		return PendingComment{}, errors.New("user id is required")
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
	kind := input.ContextKind

	if kind == "" {
		kind = "working"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO pending_comments (
			id, user_id, repo_root, context_kind, context_sha,
			file_path, diff_section, side, line_number,
			author_label, body_html, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, userID, input.RepoRoot, kind, input.ContextSHA,
		input.FilePath, input.DiffSection, input.Side, input.LineNumber,
		input.AuthorLabel, input.BodyHTML, now, now)

	if err != nil {
		return PendingComment{}, err
	}

	return PendingComment{
		ID:          id,
		UserID:      userID,
		RepoRoot:    input.RepoRoot,
		ContextKind: kind,
		ContextSHA:  input.ContextSHA,
		FilePath:    input.FilePath,
		DiffSection: input.DiffSection,
		Side:        input.Side,
		LineNumber:  input.LineNumber,
		AuthorLabel: input.AuthorLabel,
		BodyHTML:    input.BodyHTML,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) UpdatePendingComment(ctx context.Context, userID int64, id, bodyHTML string) (PendingComment, error) {
	now := s.now().UTC().Format(time.RFC3339Nano)

	result, err := s.db.ExecContext(ctx, `
		UPDATE pending_comments SET body_html = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, bodyHTML, now, id, userID)

	if err != nil {
		return PendingComment{}, err
	}

	affected, _ := result.RowsAffected()

	if affected == 0 {
		return PendingComment{}, sql.ErrNoRows
	}

	return s.GetPendingComment(ctx, userID, id)
}

func (s *Store) GetPendingComment(ctx context.Context, userID int64, id string) (PendingComment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, repo_root, context_kind, context_sha,
			file_path, diff_section, side, line_number,
			author_label, body_html, created_at, updated_at
		FROM pending_comments
		WHERE id = ? AND user_id = ?
	`, id, userID)

	return scanPendingComment(row)
}

func (s *Store) DeletePendingComment(ctx context.Context, userID int64, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM pending_comments WHERE id = ? AND user_id = ?", id, userID)

	return err
}

func (s *Store) ListPendingComments(ctx context.Context, userID int64, repoRoot, contextKind, contextSHA string) ([]PendingComment, error) {
	if contextKind == "" {
		contextKind = "working"
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, repo_root, context_kind, context_sha,
			file_path, diff_section, side, line_number,
			author_label, body_html, created_at, updated_at
		FROM pending_comments
		WHERE user_id = ? AND repo_root = ? AND context_kind = ? AND context_sha = ?
		ORDER BY file_path, line_number, created_at
	`, userID, repoRoot, contextKind, contextSHA)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []PendingComment{}

	for rows.Next() {
		comment, err := scanPendingComment(rows)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// PromotePendingComments moves every matching pending comment into
// review_comments under the given review session. Runs in a single
// transaction so a crash mid-promotion can't leave partial state.
func (s *Store) PromotePendingComments(ctx context.Context, userID int64, reviewID string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return 0, err
	}

	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		SELECT repo_root, context_kind, context_sha FROM review_sessions WHERE id = ? AND user_id = ?
	`, reviewID, userID)

	var (
		repoRoot   string
		kind       string
		contextSHA sql.NullString
	)

	if err := row.Scan(&repoRoot, &kind, &contextSHA); err != nil {
		return 0, err
	}

	now := s.now().UTC().Format(time.RFC3339Nano)

	rows, err := tx.QueryContext(ctx, `
		SELECT id, file_path, diff_section, side, line_number, author_label, body_html
		FROM pending_comments
		WHERE user_id = ? AND repo_root = ? AND context_kind = ? AND context_sha = ?
	`, userID, repoRoot, kind, fromNull(contextSHA))

	if err != nil {
		return 0, err
	}

	type promote struct {
		id, filePath, diffSection, side, authorLabel, bodyHTML string
		lineNumber                                             int64
	}

	pending := []promote{}

	for rows.Next() {
		var p promote

		if err := rows.Scan(&p.id, &p.filePath, &p.diffSection, &p.side, &p.lineNumber, &p.authorLabel, &p.bodyHTML); err != nil {
			rows.Close()

			return 0, err
		}

		pending = append(pending, p)
	}

	rows.Close()

	for _, p := range pending {
		newID := "comment-" + p.id

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO review_comments (
				id, review_id, file_path, diff_section, side, line_number,
				author_label, body_html, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, newID, reviewID, p.filePath, p.diffSection, p.side, p.lineNumber,
			p.authorLabel, p.bodyHTML, now, now); err != nil {
			return 0, err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM pending_comments WHERE id = ?`, p.id); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(pending), nil
}

func scanPendingComment(row scanner) (PendingComment, error) {
	var (
		comment    PendingComment
		contextSHA sql.NullString
	)

	if err := row.Scan(
		&comment.ID,
		&comment.UserID,
		&comment.RepoRoot,
		&comment.ContextKind,
		&contextSHA,
		&comment.FilePath,
		&comment.DiffSection,
		&comment.Side,
		&comment.LineNumber,
		&comment.AuthorLabel,
		&comment.BodyHTML,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	); err != nil {
		return PendingComment{}, err
	}

	comment.ContextSHA = fromNull(contextSHA)

	return comment, nil
}
