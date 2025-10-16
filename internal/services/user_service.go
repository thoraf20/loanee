package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/models"
	repository "github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/utils"
)

type UserService interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetLoggedInUser(ctx context.Context) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetLoggedInUser(ctx context.Context) (*models.User, error) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
			return nil, fmt.Errorf("user not found in context: %w", err)
	}

	 user, err := s.GetUserProfile(ctx, userID)
	if err != nil {
			return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}