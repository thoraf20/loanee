package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/models"
	"github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines all authentication-related operations
type AuthService interface {
	RegisterUser(ctx context.Context, input dtos.RegisterUserDTO) (*models.User, error)
	LoginUser(ctx context.Context, input dtos.LoginDTO) (string, error)
}

// authService implementation
type authService struct {
	repo repository.UserRepository
}

// NewAuthService creates a new auth service instance
func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

// RegisterUser handles user registration
func (s *authService) RegisterUser(ctx context.Context, input dtos.RegisterUserDTO) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	existingUser, _ := s.repo.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        email,
		Password: string(hash),
		IsVerified: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); 
	err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.Password = ""

	return user, nil
}

func (s *authService) LoginUser(ctx context.Context, input dtos.LoginDTO) (string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); 
	err != nil {
		fmt.Printf("warning: failed to update last login for user %s\n", user.ID.String())
	}

	// Generate JWT
	token, err := utils.GenerateToken(*user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}