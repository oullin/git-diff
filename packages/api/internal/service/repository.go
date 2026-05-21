package service

import (
	"context"
	"errors"

	"github.com/gocanto/git-diff/internal/storage"
)

// RepositoryService owns the repository-registry use cases: the list of
// known repos for a user, upserting on open, and removing. Collaborator
// management lives alongside since both surfaces share the owner-check
// contract enforced by the storage layer.
type RepositoryService struct {
	repos *storage.RepoRepo
}

func NewRepositoryService(repos *storage.RepoRepo) *RepositoryService {
	return &RepositoryService{repos: repos}
}

// ErrRepositoryPathRequired signals that a path query parameter was
// missing from a request; handlers translate it to 400.
var ErrRepositoryPathRequired = errors.New("path query parameter is required")

func (s *RepositoryService) List(
	ctx context.Context,
	userID int64,
) ([]storage.Repository, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	return s.repos.ListRepositoriesForUser(ctx, userID)
}

func (s *RepositoryService) Upsert(
	ctx context.Context,
	userID int64,
	path, name string,
) (storage.Repository, error) {
	if userID == 0 {
		return storage.Repository{}, ErrAuthenticationRequired
	}

	return s.repos.UpsertRepository(ctx, userID, path, name)
}

func (s *RepositoryService) Remove(ctx context.Context, userID int64, path string) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	if path == "" {
		return ErrRepositoryPathRequired
	}

	return s.repos.RemoveRepository(ctx, userID, path)
}

func (s *RepositoryService) ListCollaborators(
	ctx context.Context,
	userID int64,
	path string,
) ([]storage.RepositoryCollaborator, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	return s.repos.ListRepositoryCollaborators(ctx, userID, path)
}

func (s *RepositoryService) GrantCollaborator(
	ctx context.Context,
	userID int64,
	path string,
	targetUserID int64,
	role string,
) (storage.RepositoryCollaborator, error) {
	if userID == 0 {
		return storage.RepositoryCollaborator{}, ErrAuthenticationRequired
	}

	return s.repos.GrantRepositoryAccess(ctx, userID, path, targetUserID, role)
}

func (s *RepositoryService) RevokeCollaborator(
	ctx context.Context,
	userID int64,
	path string,
	targetUserID int64,
) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	return s.repos.RevokeRepositoryAccess(ctx, userID, path, targetUserID)
}
