package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

// PendingComment is a draft scoped by (user_id, repo_root, context_kind,
// context_sha); PromotePendingComments moves matching rows into review_comments.
type PendingComment struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"userId"`
	RepoRoot        string `json:"repoRoot"`
	ContextKind     string `json:"contextKind"`
	ContextSHA      string `json:"contextSha,omitempty"`
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
}

type PendingCommentInput struct {
	RepoRoot        string `json:"repoRoot"`
	ContextKind     string `json:"contextKind"`
	ContextSHA      string `json:"contextSha"`
	FilePath        string `json:"filePath"`
	DiffSection     string `json:"diffSection"`
	Side            string `json:"side"`
	LineNumber      int64  `json:"lineNumber"`
	StartLineNumber *int64 `json:"startLineNumber,omitempty"`
	StartSide       string `json:"startSide,omitempty"`
	AuthorLabel     string `json:"authorLabel"`
	BodyHTML        string `json:"bodyHtml"`
}

type PendingCommentRepo struct {
	db  *gorm.DB
	clk *clock
}

func newPendingCommentRepo(db *gorm.DB, clk *clock) *PendingCommentRepo {
	return &PendingCommentRepo{db: db, clk: clk}
}

func (r *PendingCommentRepo) CreatePendingComment(ctx context.Context, userID int64, input PendingCommentInput) (PendingComment, error) {
	if userID == 0 {
		return PendingComment{}, errors.New("user id is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	kind := input.ContextKind

	if kind == "" {
		kind = "working"
	}

	row := PendingCommentRow{
		UserID:          userID,
		RepoRoot:        input.RepoRoot,
		ContextKind:     kind,
		ContextSHA:      input.ContextSHA,
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

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return PendingComment{}, err
	}

	return toPendingComment(row), nil
}

func (r *PendingCommentRepo) UpdatePendingComment(ctx context.Context, userID int64, id int64, bodyHTML string) (PendingComment, error) {
	now := r.clk.now().UTC().Format(time.RFC3339Nano)

	result := r.db.WithContext(ctx).
		Model(&PendingCommentRow{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"body_html":  bodyHTML,
			"updated_at": now,
		})

	if result.Error != nil {
		return PendingComment{}, result.Error
	}

	if result.RowsAffected == 0 {
		return PendingComment{}, sql.ErrNoRows
	}

	return r.GetPendingComment(ctx, userID, id)
}

func (r *PendingCommentRepo) GetPendingComment(ctx context.Context, userID int64, id int64) (PendingComment, error) {
	var row PendingCommentRow

	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Take(&row).Error; err != nil {
		return PendingComment{}, err
	}

	return toPendingComment(row), nil
}

func (r *PendingCommentRepo) DeletePendingComment(ctx context.Context, userID int64, id int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&PendingCommentRow{}).Error
}

func (r *PendingCommentRepo) ListPendingComments(ctx context.Context, userID int64, repoRoot, contextKind, contextSHA string) ([]PendingComment, error) {
	if contextKind == "" {
		contextKind = "working"
	}

	var rows []PendingCommentRow

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND repo_root = ? AND context_kind = ? AND context_sha = ?", userID, repoRoot, contextKind, contextSHA).
		Order("file_path, line_number, created_at").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	comments := make([]PendingComment, 0, len(rows))

	for _, row := range rows {
		comments = append(comments, toPendingComment(row))
	}

	return comments, nil
}

// PromotePendingComments moves every pending row that matches a review's scope
// into review_comments inside a single transaction. Issues exactly three
// statements regardless of row count: one SELECT for the review scope, one
// batched INSERT, one bulk DELETE — replacing the previous per-row loop.
func (r *PendingCommentRepo) PromotePendingComments(ctx context.Context, userID int64, reviewID int64) (int, error) {
	var promoted int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var review ReviewSessionRow

		if err := tx.Select("repo_root", "context_kind", "context_sha").
			Where("id = ? AND user_id = ?", reviewID, userID).
			Take(&review).Error; err != nil {
			return err
		}

		contextSHA := ""

		if review.ContextSHA != nil {
			contextSHA = *review.ContextSHA
		}

		var pending []PendingCommentRow

		if err := tx.Where("user_id = ? AND repo_root = ? AND context_kind = ? AND context_sha = ?",
			userID, review.RepoRoot, review.ContextKind, contextSHA).
			Find(&pending).Error; err != nil {
			return err
		}

		if len(pending) == 0 {
			return nil
		}

		now := r.clk.now().UTC().Format(time.RFC3339Nano)

		comments := make([]ReviewCommentRow, 0, len(pending))
		ids := make([]int64, 0, len(pending))

		for _, p := range pending {
			comments = append(comments, ReviewCommentRow{
				ReviewID:        reviewID,
				FilePath:        p.FilePath,
				DiffSection:     p.DiffSection,
				Side:            p.Side,
				LineNumber:      p.LineNumber,
				StartLineNumber: p.StartLineNumber,
				StartSide:       p.StartSide,
				AuthorLabel:     p.AuthorLabel,
				BodyHTML:        p.BodyHTML,
				CreatedAt:       now,
				UpdatedAt:       now,
			})

			ids = append(ids, p.ID)
		}

		if err := tx.CreateInBatches(&comments, 200).Error; err != nil {
			return err
		}

		if err := tx.Where("id IN ?", ids).Delete(&PendingCommentRow{}).Error; err != nil {
			return err
		}

		promoted = len(pending)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return promoted, nil
}

func toPendingComment(row PendingCommentRow) PendingComment {
	comment := PendingComment{
		ID:              row.ID,
		UserID:          row.UserID,
		RepoRoot:        row.RepoRoot,
		ContextKind:     row.ContextKind,
		ContextSHA:      row.ContextSHA,
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

	return comment
}
