package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gocanto/git-diff/internal/storage/db"
)

type ReviewCommentInput struct {
	FilePath        string `json:"filePath"`
	DiffSection     string `json:"diffSection"`
	Side            string `json:"side"`
	LineNumber      int64  `json:"lineNumber"`
	StartLineNumber *int64 `json:"startLineNumber,omitempty"`
	StartSide       string `json:"startSide,omitempty"`
	AuthorLabel     string `json:"authorLabel"`
	BodyHTML        string `json:"bodyHtml"`
}

type ReviewComment struct {
	ID              string `json:"id"`
	ReviewID        string `json:"reviewId"`
	FilePath        string `json:"filePath"`
	DiffSection     string `json:"diffSection"`
	Side            string `json:"side"`
	LineNumber      int64  `json:"lineNumber"`
	StartLineNumber *int64 `json:"startLineNumber,omitempty"`
	StartSide       string `json:"startSide,omitempty"`
	AuthorLabel     string `json:"authorLabel"`
	BodyHTML        string `json:"bodyHtml"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	DeletedAt       string `json:"deletedAt,omitempty"`
}

// CommentRepo owns review_comments. It also writes timeline events as side
// effects via a ReviewEventWriter so comment mutations and review-event
// inserts stay consistent without coupling to the concrete ReviewRepo.
type CommentRepo struct {
	db      *sql.DB
	queries *db.Queries
	clk     *clock
	events  ReviewEventWriter
}

func newCommentRepo(conn *sql.DB, queries *db.Queries, clk *clock, events ReviewEventWriter) *CommentRepo {
	return &CommentRepo{db: conn, queries: queries, clk: clk, events: events}
}

func (r *CommentRepo) CreateReviewComment(ctx context.Context, reviewID string, commentID string, input ReviewCommentInput) (ReviewComment, error) {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	if input.AuthorLabel == "" {
		input.AuthorLabel = "You"
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO review_comments (
			id, review_id, file_path, diff_section, side, line_number,
			start_line_number, start_side,
			author_label, body_html, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, commentID, reviewID, input.FilePath, input.DiffSection, input.Side, input.LineNumber,
		nullableInt(input.StartLineNumber), nullableString(input.StartSide),
		input.AuthorLabel, input.BodyHTML, now, now)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = r.events.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_added", FilePath: input.FilePath, Message: fmt.Sprintf("Commented on line %d", input.LineNumber)})

	return ReviewComment{
		ID: commentID, ReviewID: reviewID, FilePath: input.FilePath, DiffSection: input.DiffSection,
		Side: input.Side, LineNumber: input.LineNumber,
		StartLineNumber: input.StartLineNumber, StartSide: input.StartSide,
		AuthorLabel: input.AuthorLabel, BodyHTML: input.BodyHTML, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (r *CommentRepo) UpdateReviewComment(ctx context.Context, reviewID string, commentID string, bodyHTML string) (ReviewComment, error) {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, `
		UPDATE review_comments
		SET body_html = ?, updated_at = ?, deleted_at = NULL
		WHERE id = ? AND review_id = ?
	`, bodyHTML, now, commentID, reviewID)

	if err != nil {
		return ReviewComment{}, err
	}

	comment, err := r.GetReviewComment(ctx, reviewID, commentID)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = r.events.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_edited", FilePath: comment.FilePath, Message: fmt.Sprintf("Edited comment on line %d", comment.LineNumber)})

	return comment, nil
}

func (r *CommentRepo) DeleteReviewComment(ctx context.Context, reviewID string, commentID string) error {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	comment, _ := r.GetReviewComment(ctx, reviewID, commentID)
	_, err := r.db.ExecContext(ctx, `
		UPDATE review_comments
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND review_id = ?
	`, now, now, commentID, reviewID)

	if err != nil {
		return err
	}

	if comment.ID != "" {
		_, _ = r.events.AddReviewEvent(ctx, reviewID, ReviewEventInput{Type: "comment_deleted", FilePath: comment.FilePath, Message: fmt.Sprintf("Deleted comment on line %d", comment.LineNumber)})
	}

	return nil
}

func (r *CommentRepo) GetReviewComment(ctx context.Context, reviewID string, commentID string) (ReviewComment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, start_line_number, start_side, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE id = ? AND review_id = ?
	`, commentID, reviewID)

	return scanComment(row)
}

func (r *CommentRepo) ListReviewComments(ctx context.Context, reviewID string) ([]ReviewComment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, review_id, file_path, diff_section, side, line_number, start_line_number, start_side, author_label, body_html, created_at, updated_at, deleted_at
		FROM review_comments
		WHERE review_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, reviewID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []ReviewComment{}

	for rows.Next() {
		comment, err := scanComment(rows)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func scanComment(row scanner) (ReviewComment, error) {
	var (
		comment         ReviewComment
		startLineNumber sql.NullInt64
		startSide       sql.NullString
		deletedAt       sql.NullString
	)

	if err := row.Scan(
		&comment.ID,
		&comment.ReviewID,
		&comment.FilePath,
		&comment.DiffSection,
		&comment.Side,
		&comment.LineNumber,
		&startLineNumber,
		&startSide,
		&comment.AuthorLabel,
		&comment.BodyHTML,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&deletedAt,
	); err != nil {
		return ReviewComment{}, err
	}

	if startLineNumber.Valid {
		v := startLineNumber.Int64
		comment.StartLineNumber = &v
	}

	comment.StartSide = fromNull(startSide)
	comment.DeletedAt = fromNull(deletedAt)

	return comment, nil
}
