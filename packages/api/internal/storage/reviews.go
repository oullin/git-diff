package storage

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ReviewSessionStart struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha"`
}

type ReviewSession struct {
	ID           string `json:"id"`
	RepoRoot     string `json:"repoRoot"`
	UserID       int64  `json:"userId"`
	Branch       string `json:"branch"`
	HeadSHA      string `json:"headSha"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	FilesChanged int    `json:"filesChanged"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	StartedAt    string `json:"startedAt"`
	CompletedAt  string `json:"completedAt,omitempty"`
	ContextKind  string `json:"contextKind"`
	ContextSHA   string `json:"contextSha,omitempty"`
}

type ReviewDetail struct {
	Review   ReviewSession   `json:"review"`
	Events   []ReviewEvent   `json:"events"`
	Comments []ReviewComment `json:"comments"`
}

type ReviewRepo struct {
	db     *gorm.DB
	clk    *clock
	events *ReviewEventRepo
}

func newReviewRepo(db *gorm.DB, clk *clock, events *ReviewEventRepo) *ReviewRepo {
	return &ReviewRepo{db: db, clk: clk, events: events}
}

func (r *ReviewRepo) CreateReview(ctx context.Context, userID int64, review ReviewSessionStart) (ReviewSession, error) {
	if userID == 0 {
		return ReviewSession{}, errors.New("user id is required")
	}

	now := r.clk.now().UTC().Format(time.RFC3339Nano)
	title := review.Title

	if title == "" {
		title = "Local review"
	}

	contextKind := review.ContextKind

	if contextKind == "" {
		contextKind = "working"
	}

	row := ReviewSessionRow{
		ID:           review.ID,
		RepoRoot:     review.RepoRoot,
		UserID:       userID,
		Branch:       review.Branch,
		HeadSHA:      review.HeadSHA,
		Status:       "open",
		Title:        title,
		Summary:      review.Summary,
		FilesChanged: review.FilesChanged,
		Additions:    review.Additions,
		Deletions:    review.Deletions,
		StartedAt:    now,
		ContextKind:  contextKind,
		ContextSHA:   nullableStringPtr(review.ContextSHA),
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return ReviewSession{}, err
	}

	if _, err := r.events.Add(ctx, review.ID, ReviewEventInput{
		Type:    "review_started",
		Message: title,
	}); err != nil {
		return ReviewSession{}, err
	}

	return ReviewSession{
		ID:           review.ID,
		RepoRoot:     review.RepoRoot,
		UserID:       userID,
		Branch:       review.Branch,
		HeadSHA:      review.HeadSHA,
		Status:       "open",
		Title:        title,
		Summary:      review.Summary,
		FilesChanged: review.FilesChanged,
		Additions:    review.Additions,
		Deletions:    review.Deletions,
		StartedAt:    now,
		ContextKind:  contextKind,
		ContextSHA:   review.ContextSHA,
	}, nil
}

func (r *ReviewRepo) ListReviews(ctx context.Context, userID int64, limit int64) ([]ReviewSession, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	if limit <= 0 {
		limit = 50
	}

	var rows []ReviewSessionRow

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(int(limit)).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	reviews := make([]ReviewSession, 0, len(rows))

	for _, row := range rows {
		reviews = append(reviews, toReviewSession(row))
	}

	return reviews, nil
}

// GetReviewByID returns the session without its events or comments.
func (r *ReviewRepo) GetReviewByID(ctx context.Context, id string) (ReviewSession, error) {
	var row ReviewSessionRow

	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&row).Error; err != nil {
		return ReviewSession{}, err
	}

	return toReviewSession(row), nil
}

func (r *ReviewRepo) ReviewDetail(ctx context.Context, comments *CommentRepo, id string) (ReviewDetail, error) {
	review, err := r.GetReviewByID(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	events, err := r.events.List(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	commentList, err := comments.ListReviewComments(ctx, id)

	if err != nil {
		return ReviewDetail{}, err
	}

	return ReviewDetail{Review: review, Events: events, Comments: commentList}, nil
}

func toReviewSession(row ReviewSessionRow) ReviewSession {
	review := ReviewSession{
		ID:           row.ID,
		RepoRoot:     row.RepoRoot,
		UserID:       row.UserID,
		Branch:       row.Branch,
		HeadSHA:      row.HeadSHA,
		Status:       row.Status,
		Title:        row.Title,
		Summary:      row.Summary,
		FilesChanged: row.FilesChanged,
		Additions:    row.Additions,
		Deletions:    row.Deletions,
		StartedAt:    row.StartedAt,
		ContextKind:  row.ContextKind,
	}

	if row.CompletedAt != nil {
		review.CompletedAt = *row.CompletedAt
	}

	if row.ContextSHA != nil {
		review.ContextSHA = *row.ContextSHA
	}

	return review
}

func nullableStringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
