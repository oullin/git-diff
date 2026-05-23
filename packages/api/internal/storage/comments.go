package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/gocanto/git-diff/internal/db"
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
	ID              int64  `json:"id"`
	ReviewID        int64  `json:"reviewId"`
	FilePath        string `json:"filePath"`
	DiffSection     string `json:"diffSection"`
	Side            string `json:"side"`
	LineNumber      int64  `json:"lineNumber"`
	StartLineNumber *int64 `json:"startLineNumber,omitempty"`
	StartSide       string `json:"startSide,omitempty"`
	AuthorLabel     string `json:"authorLabel"`
	BodyHTML        string `json:"bodyHtml"`
	DeletedAt       string `json:"deletedAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type CommentRepo struct {
	events *ReviewEventRepo
}

func newCommentRepo(events *ReviewEventRepo) *CommentRepo {
	return &CommentRepo{events: events}
}

func (r *CommentRepo) CreateReviewComment(ctx context.Context, reviewID int64, input ReviewCommentInput) (ReviewComment, error) {
	now := db.Now().UTC().Format(time.RFC3339Nano)

	if input.AuthorLabel == "" {
		input.AuthorLabel = "You"
	}

	row := ReviewCommentRow{
		ReviewID:        reviewID,
		FilePath:        input.FilePath,
		DiffSection:     input.DiffSection,
		Side:            input.Side,
		LineNumber:      input.LineNumber,
		StartLineNumber: input.StartLineNumber,
		StartSide:       nullableStringPtr(input.StartSide),
		AuthorLabel:     input.AuthorLabel,
		BodyHTML:        input.BodyHTML,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := db.Conn().WithContext(ctx).Create(&row).Error; err != nil {
		return ReviewComment{}, err
	}

	_, _ = r.events.Add(ctx, reviewID, ReviewEventInput{Type: "comment_added", FilePath: input.FilePath, Message: fmt.Sprintf("Commented on line %d", input.LineNumber)})

	return ReviewComment{
		ID:              row.ID,
		ReviewID:        reviewID,
		FilePath:        input.FilePath,
		DiffSection:     input.DiffSection,
		Side:            input.Side,
		LineNumber:      input.LineNumber,
		StartLineNumber: input.StartLineNumber,
		StartSide:       input.StartSide,
		AuthorLabel:     input.AuthorLabel,
		BodyHTML:        input.BodyHTML,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *CommentRepo) UpdateReviewComment(ctx context.Context, reviewID int64, commentID int64, bodyHTML string) (ReviewComment, error) {
	now := db.Now().UTC().Format(time.RFC3339Nano)

	if err := db.Conn().WithContext(ctx).
		Model(&ReviewCommentRow{}).
		Where("id = ? AND review_id = ?", commentID, reviewID).
		Updates(map[string]any{
			"body_html":  bodyHTML,
			"updated_at": now,
			"deleted_at": nil,
		}).Error; err != nil {
		return ReviewComment{}, err
	}

	comment, err := r.GetReviewComment(ctx, reviewID, commentID)

	if err != nil {
		return ReviewComment{}, err
	}

	_, _ = r.events.Add(ctx, reviewID, ReviewEventInput{Type: "comment_edited", FilePath: comment.FilePath, Message: fmt.Sprintf("Edited comment on line %d", comment.LineNumber)})

	return comment, nil
}

func (r *CommentRepo) DeleteReviewComment(ctx context.Context, reviewID int64, commentID int64) error {
	now := db.Now().UTC().Format(time.RFC3339Nano)
	comment, _ := r.GetReviewComment(ctx, reviewID, commentID)

	if err := db.Conn().WithContext(ctx).
		Model(&ReviewCommentRow{}).
		Where("id = ? AND review_id = ?", commentID, reviewID).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	if comment.ID != 0 {
		_, _ = r.events.Add(ctx, reviewID, ReviewEventInput{Type: "comment_deleted", FilePath: comment.FilePath, Message: fmt.Sprintf("Deleted comment on line %d", comment.LineNumber)})
	}

	return nil
}

func (r *CommentRepo) GetReviewComment(ctx context.Context, reviewID int64, commentID int64) (ReviewComment, error) {
	var row ReviewCommentRow

	if err := db.Conn().WithContext(ctx).
		Where("id = ? AND review_id = ?", commentID, reviewID).
		Take(&row).Error; err != nil {
		return ReviewComment{}, err
	}

	return toReviewComment(row), nil
}

func (r *CommentRepo) ListReviewComments(ctx context.Context, reviewID int64) ([]ReviewComment, error) {
	var rows []ReviewCommentRow

	if err := db.Conn().WithContext(ctx).
		Where("review_id = ? AND deleted_at IS NULL", reviewID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	comments := make([]ReviewComment, 0, len(rows))

	for _, row := range rows {
		comments = append(comments, toReviewComment(row))
	}

	return comments, nil
}

func toReviewComment(row ReviewCommentRow) ReviewComment {
	comment := ReviewComment{
		ID:              row.ID,
		ReviewID:        row.ReviewID,
		FilePath:        row.FilePath,
		DiffSection:     row.DiffSection,
		Side:            row.Side,
		LineNumber:      row.LineNumber,
		StartLineNumber: row.StartLineNumber,
		AuthorLabel:     row.AuthorLabel,
		BodyHTML:        row.BodyHTML,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}

	if row.StartSide != nil {
		comment.StartSide = *row.StartSide
	}

	if row.DeletedAt != nil {
		comment.DeletedAt = *row.DeletedAt
	}

	return comment
}
