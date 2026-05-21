package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gocanto/git-diff/internal/storage"
)

// ReviewService owns review-session, review-event, and review-comment use
// cases. It abstracts away the store handle, the random-ID generation, and
// the auto-event side effects so handlers can stay thin (decode -> dispatch
// -> writeJSON).
type ReviewService struct {
	store *storage.Store
}

func NewReviewService(store *storage.Store) *ReviewService {
	return &ReviewService{store: store}
}

// ErrAuthenticationRequired signals that an action needs a logged-in user;
// handlers translate it to 401.
var ErrAuthenticationRequired = errors.New("authentication required")

// Create starts a new review session for `userID`. An empty input.ID is
// auto-filled with a random identifier so callers can choose to delegate
// id generation here.
func (s *ReviewService) Create(
	ctx context.Context,
	userID int64,
	input storage.ReviewSessionStart,
) (storage.ReviewSession, error) {
	if userID == 0 {
		return storage.ReviewSession{}, ErrAuthenticationRequired
	}

	if input.ID == "" {
		input.ID = randomID("review")
	}

	return s.store.CreateReview(ctx, userID, input)
}

// List returns up to `limit` review sessions owned by `userID`, most
// recent first.
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

	return s.store.ListReviews(ctx, userID, limit)
}

// Detail returns the session, its events, and its comments by review id.
func (s *ReviewService) Detail(ctx context.Context, id string) (storage.ReviewDetail, error) {
	return s.store.ReviewDetail(ctx, id)
}

// AddEvent appends a review-timeline event to the session.
func (s *ReviewService) AddEvent(
	ctx context.Context,
	reviewID string,
	input storage.ReviewEventInput,
) (storage.ReviewEvent, error) {
	return s.store.AddReviewEvent(ctx, reviewID, input)
}

// CreateComment posts a new comment under the review session and
// auto-generates its identifier. The store records a "comment_added"
// review event side-effect.
func (s *ReviewService) CreateComment(
	ctx context.Context,
	reviewID string,
	input storage.ReviewCommentInput,
) (storage.ReviewComment, error) {
	return s.store.CreateReviewComment(ctx, reviewID, randomID("comment"), input)
}

// UpdateComment rewrites the body of an existing comment and emits a
// "comment_edited" event.
func (s *ReviewService) UpdateComment(
	ctx context.Context,
	reviewID, commentID, bodyHTML string,
) (storage.ReviewComment, error) {
	return s.store.UpdateReviewComment(ctx, reviewID, commentID, bodyHTML)
}

// DeleteComment soft-deletes the comment and emits a "comment_deleted" event.
func (s *ReviewService) DeleteComment(ctx context.Context, reviewID, commentID string) error {
	return s.store.DeleteReviewComment(ctx, reviewID, commentID)
}

func randomID(prefix string) string {
	var bytes [12]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s-fallback", prefix)
	}

	return prefix + "-" + hex.EncodeToString(bytes[:])
}
