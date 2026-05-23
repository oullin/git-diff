package service

import (
	"context"

	"github.com/gocanto/git-diff/internal/storage"
)

type CollaboratorService struct {
	collaborators *storage.CollaboratorRepo
}

func NewCollaboratorService(collaborators *storage.CollaboratorRepo) *CollaboratorService {
	return &CollaboratorService{collaborators: collaborators}
}

func (s *CollaboratorService) List(
	ctx context.Context,
	userID int64,
	path string,
) ([]storage.RepositoryCollaborator, error) {
	if userID == 0 {
		return nil, ErrAuthenticationRequired
	}

	return s.collaborators.List(ctx, userID, path)
}

func (s *CollaboratorService) Grant(
	ctx context.Context,
	userID int64,
	path string,
	targetUserID int64,
	role string,
) (storage.RepositoryCollaborator, error) {
	if userID == 0 {
		return storage.RepositoryCollaborator{}, ErrAuthenticationRequired
	}

	return s.collaborators.Grant(ctx, userID, path, targetUserID, role)
}

func (s *CollaboratorService) Revoke(
	ctx context.Context,
	userID int64,
	path string,
	targetUserID int64,
) error {
	if userID == 0 {
		return ErrAuthenticationRequired
	}

	return s.collaborators.Revoke(ctx, userID, path, targetUserID)
}
