package service

import (
	"context"
	"errors"
	"strings"

	"github.com/gocanto/git-diff/internal/storage"
)

type PendingCommentService struct {
	pending *storage.PendingCommentRepo
}

func NewPendingCommentService(pending *storage.PendingCommentRepo) *PendingCommentService {
	return &PendingCommentService{pending: pending}
}

var ErrInvalidCommentInput = errors.New("filePath and diffSection are required")

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

	return s.pending.CreatePendingComment(ctx, userID, input)
}

func (s *PendingCommentService) Update(
	ctx context.Context,
	userID int64,
	id int64,
	bodyHTML string,
) (storage.PendingComment, error) {
	if userID == 0 {
		return storage.PendingComment{}, ErrAuthenticationRequired
	}

	return s.pending.UpdatePendingComment(ctx, userID, id, bodyHTML)
}

func (s *PendingCommentService) Delete(ctx context.Context, userID int64, id int64) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	return s.pending.DeletePendingComment(ctx, userID, id)
}

func (s *PendingCommentService) Promote(
	ctx context.Context,
	userID int64,
	reviewID int64,
) (int, error) {
	if userID == 0 {
		return 0, ErrAuthenticationRequired
	}

	if reviewID == 0 {
		return 0, ErrReviewIDRequired
	}

	return s.pending.PromotePendingComments(ctx, userID, reviewID)
}
