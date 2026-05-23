package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/gocanto/git-diff/internal/review"
	"github.com/gocanto/git-diff/internal/storage"
)

type BranchService struct {
	branches *storage.RepositoryBranchRepo
	repos    *storage.RepoRepo
}

// ListResult carries git branch names; Records is populated only when the
// store sync succeeded.
type ListResult struct {
	Names   []string         `json:"branches"`
	Records []storage.Branch `json:"records,omitempty"`
}

func NewBranchService(branches *storage.RepositoryBranchRepo, repos *storage.RepoRepo) *BranchService {
	return &BranchService{branches: branches, repos: repos}
}

var ErrBranchNameRequired = errors.New("name is required")

var ErrBranchAlreadyLocked = errors.New("branch is locked")

// resolveRepositoryID maps a git root path to its numeric repositories.id.
// Returns (0, nil) when the repo isn't registered — callers degrade to the
// git-side data only.
func (s *BranchService) resolveRepositoryID(ctx context.Context, path string) (int64, error) {
	root, err := review.ResolveRoot(ctx, path)

	if err != nil {
		return 0, err
	}

	repo, err := s.repos.GetByPath(ctx, root)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return repo.ID, nil
}

// List degrades gracefully: if the store sync fails the result still
// contains the git-side names so the UI keeps working.
func (s *BranchService) List(ctx context.Context, path string) (ListResult, error) {
	names, err := review.ListBranches(ctx, path)

	if err != nil {
		return ListResult{}, err
	}

	repoID, err := s.resolveRepositoryID(ctx, path)

	if err != nil || repoID == 0 {
		return ListResult{Names: names}, nil
	}

	_ = s.branches.SyncBranches(ctx, repoID, names)

	records, err := s.branches.ListBranches(ctx, repoID)

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

	repoID, err := s.resolveRepositoryID(ctx, path)

	if err != nil {
		return nil, err
	}

	if repoID == 0 {
		return nil, storage.ErrRepositoryNotFound
	}

	if err := s.branches.LockBranch(ctx, repoID, name, userID); err != nil {
		return nil, err
	}

	return s.branches.ListBranches(ctx, repoID)
}

func (s *BranchService) Unlock(
	ctx context.Context,
	userID int64,
	path, name string,
) ([]storage.Branch, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	repoID, err := s.resolveRepositoryID(ctx, path)

	if err != nil {
		return nil, err
	}

	if repoID == 0 {
		return nil, storage.ErrRepositoryNotFound
	}

	if err := s.branches.UnlockBranch(ctx, repoID, name); err != nil {
		return nil, err
	}

	return s.branches.ListBranches(ctx, repoID)
}

func (s *BranchService) Delete(ctx context.Context, userID int64, path, name string) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	if name == "" {
		return ErrBranchNameRequired
	}

	repoID, err := s.resolveRepositoryID(ctx, path)

	if err != nil {
		return err
	}

	if repoID != 0 {
		locked, err := s.branches.IsBranchLocked(ctx, repoID, name)

		if err != nil {
			return err
		}

		if locked {
			return ErrBranchAlreadyLocked
		}
	}

	if err := review.DeleteBranch(ctx, path, name); err != nil {
		return err
	}

	if repoID != 0 {
		_ = s.branches.DeleteBranchRow(ctx, repoID, name)
	}

	return nil
}
