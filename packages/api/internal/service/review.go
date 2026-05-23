package service

import (
	"context"
	"errors"

	"github.com/gocanto/git-diff/internal/storage"
)

type ReviewService struct {
	reviews  *storage.ReviewRepo
	events   *storage.ReviewEventRepo
	comments *storage.CommentRepo
}

func NewReviewService(reviews *storage.ReviewRepo, events *storage.ReviewEventRepo, comments *storage.CommentRepo) *ReviewService {
	return &ReviewService{reviews: reviews, events: events, comments: comments}
}

var ErrAuthenticationRequired = errors.New("authentication required")

func (s *ReviewService) Create(
	ctx context.Context,
	userID int64,
	input storage.ReviewSessionStart,
) (storage.ReviewSession, error) {
	if userID == 0 {
		return storage.ReviewSession{}, ErrAuthenticationRequired
	}

	return s.reviews.CreateReview(ctx, userID, input)
}

func (s *ReviewService) List(
	ctx context.Context,
	userID int64,
	limit int64,
) ([]storage.ReviewSession, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	if limit <= 0 {
		limit = 50
	}

	return s.reviews.ListReviews(ctx, userID, limit)
}

func (s *ReviewService) Detail(ctx context.Context, id int64) (storage.ReviewDetail, error) {
	return s.reviews.ReviewDetail(ctx, s.comments, id)
}

func (s *ReviewService) AddEvent(
	ctx context.Context,
	reviewID int64,
	input storage.ReviewEventInput,
) (storage.ReviewEvent, error) {
	return s.events.Add(ctx, reviewID, input)
}

// CreateComment records a "comment_added" review event as a side effect.
func (s *ReviewService) CreateComment(
	ctx context.Context,
	reviewID int64,
	input storage.ReviewCommentInput,
) (storage.ReviewComment, error) {
	return s.comments.CreateReviewComment(ctx, reviewID, input)
}

// UpdateComment emits a "comment_edited" event.
func (s *ReviewService) UpdateComment(
	ctx context.Context,
	reviewID, commentID int64,
	bodyHTML string,
) (storage.ReviewComment, error) {
	return s.comments.UpdateReviewComment(ctx, reviewID, commentID, bodyHTML)
}

// DeleteComment soft-deletes the comment and emits a "comment_deleted" event.
func (s *ReviewService) DeleteComment(ctx context.Context, reviewID, commentID int64) error {
	return s.comments.DeleteReviewComment(ctx, reviewID, commentID)
}
