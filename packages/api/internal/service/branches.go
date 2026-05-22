package service

import (
	"context"
	"errors"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
)

type BranchService struct {
	branches *storage.BranchRepo
}

// ListResult carries git branch names; Records is populated only when the
// store sync succeeded.
type ListResult struct {
	Names   []string         `json:"branches"`
	Records []storage.Branch `json:"records,omitempty"`
}

func NewBranchService(branches *storage.BranchRepo) *BranchService {
	return &BranchService{branches: branches}
}

var ErrBranchNameRequired = errors.New("name is required")

var ErrBranchAlreadyLocked = errors.New("branch is locked")

// List degrades gracefully: if the store sync fails the result still
// contains the git-side names so the UI keeps working.
func (s *BranchService) List(ctx context.Context, path string) (ListResult, error) {
	names, err := review.ListBranches(ctx, path)

	if err != nil {
		return ListResult{}, err
	}

	root, err := review.ResolveRoot(ctx, path)

	if err != nil {
		return ListResult{Names: names}, nil
	}

	_ = s.branches.SyncBranches(ctx, root, names)

	records, err := s.branches.ListBranches(ctx, root)

	if err != nil {
		return ListResult{Names: names}, nil
	}

	return ListResult{Names: names, Records: records}, nil
}

func (s *BranchService) Lock(
	ctx context.Context,
	userID int64,
	path, name string,
) ([]storage.Branch, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	root, err := review.ResolveRoot(ctx, path)

	if err != nil {
		return nil, err
	}

	if err := s.branches.LockBranch(ctx, root, name, userID); err != nil {
		return nil, err
	}

	return s.branches.ListBranches(ctx, root)
}

func (s *BranchService) Unlock(
	ctx context.Context,
	userID int64,
	path, name string,
) ([]storage.Branch, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	root, err := review.ResolveRoot(ctx, path)

	if err != nil {
		return nil, err
	}

	if err := s.branches.UnlockBranch(ctx, root, name); err != nil {
		return nil, err
	}

	return s.branches.ListBranches(ctx, root)
}

func (s *BranchService) Delete(ctx context.Context, userID int64, path, name string) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	if name == "" {
		return ErrBranchNameRequired
	}

	root, err := review.ResolveRoot(ctx, path)

	if err != nil {
		return err
	}

	locked, err := s.branches.IsBranchLocked(ctx, root, name)

	if err != nil {
		return err
	}

	if locked {
		return ErrBranchAlreadyLocked
	}

	if err := review.DeleteBranch(ctx, path, name); err != nil {
		return err
	}

	_ = s.branches.DeleteBranchRow(ctx, root, name)

	return nil
}
