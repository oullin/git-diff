package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/gocanto/git-diff/internal/storage"
)

// PendingCommentService owns the draft-comment use cases. The promote
// path is transactional inside the storage layer; the service just
// proxies to it but normalises validation + id generation so handlers
// stop hand-rolling them.
type PendingCommentService struct {
	pending *storage.PendingCommentRepo
}

func NewPendingCommentService(pending *storage.PendingCommentRepo) *PendingCommentService {
	return &PendingCommentService{pending: pending}
}

// ErrInvalidCommentInput signals that filePath or diffSection were missing
// from a create request; handlers translate it to 400.
var ErrInvalidCommentInput = errors.New("filePath and diffSection are required")

// ErrReviewIDRequired signals that the promote request omitted the review
// id; handlers translate it to 400.
var ErrReviewIDRequired = errors.New("reviewId is required")

func (s *PendingCommentService) List(
	ctx context.Context,
	userID int64,
	repoRoot, kind, sha string,
) ([]storage.PendingComment, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	return s.pending.ListPendingComments(ctx, userID, repoRoot, kind, sha)
}

func (s *PendingCommentService) Create(
	ctx context.Context,
	userID int64,
	defaultRepoRoot string,
	input storage.PendingCommentInput,
) (storage.PendingComment, error) {
	if userID == 0 {
		return storage.PendingComment{}, ErrAuthenticationRequired
	}

	if strings.TrimSpace(input.RepoRoot) == "" {
		input.RepoRoot = defaultRepoRoot
	}

	if strings.TrimSpace(input.FilePath) == "" || strings.TrimSpace(input.DiffSection) == "" {
		return storage.PendingComment{}, ErrInvalidCommentInput
	}

	return s.pending.CreatePendingComment(ctx, userID, "pending-"+randID(8), input)
}

func (s *PendingCommentService) Update(
	ctx context.Context,
	userID int64,
	id, bodyHTML string,
) (storage.PendingComment, error) {
	if userID == 0 {
		return storage.PendingComment{}, ErrAuthenticationRequired
	}

	return s.pending.UpdatePendingComment(ctx, userID, id, bodyHTML)
}

func (s *PendingCommentService) Delete(ctx context.Context, userID int64, id string) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	return s.pending.DeletePendingComment(ctx, userID, id)
}

func (s *PendingCommentService) Promote(
	ctx context.Context,
	userID int64,
	reviewID string,
) (int, error) {
	if userID == 0 {
		return 0, ErrAuthenticationRequired
	}

	if strings.TrimSpace(reviewID) == "" {
		return 0, ErrReviewIDRequired
	}

	return s.pending.PromotePendingComments(ctx, userID, reviewID)
}

func randID(n int) string {
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		return "0"
	}

	return hex.EncodeToString(buf)
}
