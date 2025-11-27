package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Service struct {
	repo   Repository
	logger zerolog.Logger

}

func NewService(repo Repository, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
  return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
  return s.repo.GetByEmail(ctx, email)
}

func (s *Service) Update(ctx context.Context, user *User) error {
  return s.repo.Update(ctx, user)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*User, int64, error) {
  return s.repo.List(ctx, limit, offset)
}

// ChangePassword changes user password (authenticated user)
// func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
// 	s.logger.Info().Any("user_id", userID).Msg("Password change attempt")

// 	user, err := s.repo.GetByID(ctx, userID)
// 	if err != nil {
// 		return err
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
// 		s.logger.Warn().Any("user_id", userID).Msg("Invalid old password")
// 		return e.NewBadRequestError("invalid old password")
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("Failed to hash new password")
// 		return e.NewInternalError("failed to hash password", err)
// 	}

// 	user.Password = string(hashedPassword)

// 	if err := s.repo.Update(ctx, user); err != nil {
// 		return err
// 	}

// 	s.logger.Info().Any("user_id", userID).Msg("Password changed successfully")
// 	return nil
// }

// // RefreshToken generates a new access token from refresh token
// func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
// 	s.logger.Info().Msg("Token refresh attempt")

// 	// Validate refresh token
// 	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Verify user still exists and is active
// 	user, err := s.repo.GetUserByID(ctx, claims.UserID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Generate new access token
// 	accessToken, err := s.jwtManager.GenerateAccessToken(
// 		user.ID,
// 		user.Email,
// 		user.FirstName+" "+user.LastName,
// 		user.Role,
// 	)
// 	if err != nil {
// 		s.logger.Error().Err(err).Any("user_id", user.ID).Msg("Failed to generate access token")
// 		return nil, err
// 	}

// 	s.logger.Info().
// 		Any("user_id", user.ID).
// 		Msg("Token refreshed successfully")

// 	return &RefreshTokenResponse{
// 		AccessToken: accessToken,
// 	}, nil
// }