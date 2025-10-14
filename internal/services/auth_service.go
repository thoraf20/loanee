package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/thoraf20/loanee/internal/cache"
	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/models"
	"github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines all authentication-related operations
type AuthService interface {
	RegisterUser(ctx context.Context, input dtos.RegisterUserDTO) (*models.User, error)
	VerifyEmail(ctx context.Context, input dtos.VerifyEmailDTO) (map[string]string, error)
	LoginUser(ctx context.Context, input dtos.LoginDTO) (string, error)
	PasswordResetRequest(ctx context.Context, input dtos.PasswordRequestDTO) (map[string]string, error)
	PasswordReset(ctx context.Context, input dtos.PasswordResetDTO) (map[string]string, error)
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

	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
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

	code := utils.GenerateCode(6)

	fmt.Print(code, "code")

	// Store code in Redis with TTL (10 minutes)
	cacheKey := fmt.Sprintf("email-verification-%s", email)
	cache.CacheSet(cacheKey, code, 10*time.Minute)

	return user, nil
}

func (s *authService) VerifyEmail(ctx context.Context, input dtos.VerifyEmailDTO) (map[string]string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	cacheKey := fmt.Sprintf("email-verification-%s", email)
	storedCode, err := cache.CacheGet(cacheKey)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("reset code expired or not found")
		}
		return nil, fmt.Errorf("error fetching reset code: %v", err)
	}

	if storedCode != input.Code {
		return nil, errors.New("invalid reset code")
	}

	user.IsVerified = true
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	cache.Redis.Del(cache.Ctx, cacheKey)

	return map[string]string{
		"message": "email verification successful",
	}, nil
}

func (s *authService) LoginUser(ctx context.Context, input dtos.LoginDTO) (string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)

	if err != nil {
		return "", errors.New("invalid credentials")
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
func (s *authService) PasswordResetRequest(ctx context.Context, input dtos.PasswordRequestDTO) (map[string]string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	code := utils.GenerateCode(6)

	cacheKey := fmt.Sprintf("password-reset-%s", email)
	cache.CacheSet(cacheKey, code, 10*time.Minute)

	return map[string]string{
		"message":    "password reset code generated",
		"reset_code": code, //remove this in production
	}, nil
}

func (s *authService) PasswordReset(ctx context.Context, input dtos.PasswordResetDTO) (map[string]string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	cacheKey := fmt.Sprintf("password-reset-%s", email)
	storedCode, err := cache.CacheGet(cacheKey)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("reset code expired or not found")
		}
		return nil, fmt.Errorf("error fetching reset code: %v", err)
	}

	if storedCode != input.Code {
		return nil, errors.New("invalid reset code")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hash)
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	cache.Redis.Del(cache.Ctx, cacheKey)

	return map[string]string{
		"message": "password reset successful",
	}, nil
}
