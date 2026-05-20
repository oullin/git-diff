package service

import (
	"context"
	"errors"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
)

// BranchService owns the branch-management use cases that combine git
// state (resolve repo root, delete branch via git) with the DB-backed
// lock table. The service hides the two-step coordination and the
// resolve-then-act pattern that every handler was duplicating.
type BranchService struct {
	store *storage.Store
}

func NewBranchService(store *storage.Store) *BranchService {
	return &BranchService{store: store}
}

// ErrBranchNameRequired signals that a delete or lock request omitted the
// branch name; handlers translate it to 400.
var ErrBranchNameRequired = errors.New("name is required")

// ErrBranchAlreadyLocked signals that a delete was refused because the
// branch is locked; handlers translate it to 409.
var ErrBranchAlreadyLocked = errors.New("branch is locked")

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

	if err := s.store.LockBranch(ctx, root, name, userID); err != nil {
		return nil, err
	}

	return s.store.ListBranches(ctx, root)
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

	if err := s.store.UnlockBranch(ctx, root, name); err != nil {
		return nil, err
	}

	return s.store.ListBranches(ctx, root)
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

	locked, err := s.store.IsBranchLocked(ctx, root, name)

	if err != nil {
		return err
	}

	if locked {
		return ErrBranchAlreadyLocked
	}

	if err := review.DeleteBranch(ctx, path, name); err != nil {
		return err
	}

	_ = s.store.DeleteBranchRow(ctx, root, name)

	return nil
}
