package service

import (
	"context"
	"errors"

	"github.com/oullin/git-diff/internal/storage"
)

type RepositoryService struct {
	repos *storage.RepoRepo
}

func NewRepositoryService(repos *storage.RepoRepo) *RepositoryService {
	return &RepositoryService{repos: repos}
}

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
