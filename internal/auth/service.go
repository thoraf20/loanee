package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/user"
	jwt "github.com/thoraf20/loanee/internal/utils"
	e "github.com/thoraf20/loanee/pkg/error"
	"github.com/thoraf20/loanee/pkg/tokenblacklist"
)

type Service struct {
	repo       Repository
	jwtManager *jwt.Manager
	tokenBlacklist tokenblacklist.Blacklist
	config     *config.Config
	logger     zerolog.Logger
}

func NewService(
	repo Repository,
	jwtManager *jwt.Manager,
	tokenBlacklist tokenblacklist.Blacklist,
	config *config.Config,
	logger zerolog.Logger,
) *Service {
	return &Service{
		repo:       repo,
		jwtManager: jwtManager,
		tokenBlacklist: tokenBlacklist,
		config:     config,
		logger:     logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	s.logger.Info().Str("email", req.Email).Msg("Registering new user")

	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, e.ErrUserEmailExists
	}

	// Check if error is something other than not found
	if err != nil && !e.IsAppError(err) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, e.NewInternalError("failed to hash password", err)
	}

	// Create user
	newUser := &user.User{
		Email:         req.Email,
		Password:      string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		PhoneNumber:   &req.Phone,
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to create user")
		return nil, err
	}

	// Generate verification code
	verificationCode, err := s.generateVerificationCode()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate verification code")
		// Don't fail registration if verification code generation fails
	} else {
		// Save verification code
		expiresAt := time.Now().Add(24 * time.Hour)
		if err := s.repo.SaveVerificationCode(ctx, newUser.ID, verificationCode, expiresAt); err != nil {
			s.logger.Error().Err(err).Msg("Failed to save verification code")
		} else {
			// TODO: Send verification email
			s.logger.Info().
				Any("user_id", newUser.ID).
				Str("code", verificationCode).
				Msg("Verification code generated (TODO: send email)")
		}
	}

	s.logger.Info().
		Any("user_id", newUser.ID).
		Str("email", newUser.Email).
		Str("code", verificationCode).
		Msg("User registered successfully")

	return &RegisterResponse{
		User:    newUser,
		Message: "Registration successful. Please check your email for verification code.",
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	s.logger.Info().Str("email", req.Email).Msg("User login attempt")

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if appErr := e.GetAppError(err); appErr != nil && appErr.Code == e.CodeNotFound {
			return nil, e.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn().
			Any("user_id", user.ID).
			Str("email", req.Email).
			Msg("Invalid password attempt")
		return nil, e.ErrInvalidCredentials
	}

	// Check if email is verified (optional based on your requirements)
	if !user.IsVerified  && s.config.App.RequireEmailVerification {
		return nil, e.ErrEmailNotVerified
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		user.Email,
		user.FirstName+" "+user.LastName,
		user.Role,
	)
	if err != nil {
		s.logger.Error().Err(err).Any("user_id", user.ID).Msg("Failed to generate tokens")
		return nil, err
	}

	s.logger.Info().
		Any("user_id", user.ID).
		Str("email", user.Email).
		Msg("User logged in successfully")

	// Remove password from response
	user.Password = ""

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	s.logger.Info().Str("email", req.Email).Msg("Verifying email")

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if user.IsVerified {
		return nil, e.NewBadRequestError("email is already verified")
	}

	verificationCode, err := s.repo.GetVerificationCode(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Verify code
	if err := bcrypt.CompareHashAndPassword([]byte(verificationCode.Code), []byte(req.Code)); err != nil {
		s.logger.Warn().
			Any("user_id", user.ID).
			Msg("Invalid verification code attempt")
		return nil, e.ErrVerificationCodeInvalid
	}

	// Check expiration
	if time.Now().After(verificationCode.ExpiresAt) {
		return nil, e.ErrVerificationCodeExpired
	}

	// Mark email as verified
	if err := s.repo.MarkEmailAsVerified(ctx, user.ID); err != nil {
		s.logger.Error().Err(err).Any("user_id", user.ID).Msg("Failed to mark email as verified")
		return nil, err
	}

	// Mark code as used (update verification code)
	if err := s.repo.InvalidateVerificationCode(ctx, verificationCode.ID, req.Code); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to invalidate verification code")
	}

	s.logger.Info().
		Any("user_id", user.ID).
		Str("email", user.Email).
		Msg("Email verified successfully")

	return &VerifyEmailResponse{
		Message: "Email verified successfully",
		Verified: true,
	}, nil
}

// ResendVerificationCode generates and sends a new verification code
func (s *Service) ResendVerificationCode(ctx context.Context, req *ResendCodeRequest) error {
	s.logger.Info().Str("email", req.Email).Msg("Resending verification code")

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	// Check if already verified
	if user.IsVerified {
		return e.NewBadRequestError("email is already verified")
	}

	// Generate new verification code
	verificationCode, err := s.generateVerificationCode()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate verification code")
		return e.NewInternalError("failed to generate verification code", err)
	}

	// Save verification code
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.SaveVerificationCode(ctx, user.ID, verificationCode, expiresAt); err != nil {
		return err
	}

	// TODO: Send verification email
	s.logger.Info().
		Any("user_id", user.ID).
		Str("code", verificationCode).
		Msg("Verification code generated (TODO: send email)")

	return nil
}

// ForgotPassword initiates password reset process
func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	s.logger.Info().Str("email", req.Email).Msg("Password reset requested")

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if appErr := e.GetAppError(err); appErr != nil && appErr.Code == e.CodeNotFound {
			s.logger.Info().Str("email", req.Email).Msg("Password reset requested for non-existent user")
			return nil
		}
		return err
	}

	resetToken, err := s.generateResetToken()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate reset token")
		return e.NewInternalError("failed to generate reset token", err)
	}

	// Save reset token
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.repo.SavePasswordResetToken(ctx, user.ID, resetToken, expiresAt); err != nil {
		return err
	}

	// TODO: Send password reset email
	s.logger.Info().
		Any("user_id", user.ID).
		Str("token", resetToken).
		Msg("Password reset token generated (TODO: send email)")

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	s.logger.Info().Msg("Password reset attempt")

	resetToken, err := s.repo.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return err
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return e.ErrPasswordResetTokenNotFound.WithDetail("reason", "token expired")
	}

	if resetToken.Used {
		return e.ErrPasswordResetTokenNotFound.WithDetail("reason", "token already used")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash new password")
		return e.NewInternalError("failed to hash password", err)
	}

	if err := s.repo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return err
	}

	if err := s.repo.InvalidatePasswordResetToken(ctx, req.Token); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to invalidate reset token")
	}

	s.logger.Info().
		Any("user_id", resetToken.UserID).
		Msg("Password reset successfully")

	return nil
}

// Logout invalidates user tokens by adding them to blacklist
func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	s.logger.Info().Msg("User logout initiated")

	// Validate and extract claims from access token
	accessClaims, err := s.jwtManager.ValidateToken(accessToken)
	if err != nil {
		// Token might be invalid or expired, but we still try to blacklist
		s.logger.Warn().Err(err).Msg("Invalid access token during logout")
	}

	// Validate and extract claims from refresh token
	refreshClaims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Invalid refresh token during logout")
	}

	// Blacklist access token
	if accessClaims != nil {
		if err := s.tokenBlacklist.Add(ctx, accessToken, accessClaims.ExpiresAt.Time); err != nil {
			s.logger.Error().Err(err).Msg("Failed to blacklist access token")
			return err
		}
	}

	// Blacklist refresh token
	if refreshClaims != nil {
		if err := s.tokenBlacklist.Add(ctx, refreshToken, refreshClaims.ExpiresAt.Time); err != nil {
			s.logger.Error().Err(err).Msg("Failed to blacklist refresh token")
			return err
		}
	}

	s.logger.Info().
		Any("user_id", accessClaims.UserID).
		Msg("User logged out successfully")

	return nil
}

// LogoutAllDevices logs out user from all devices (blacklist all active tokens)
func (s *Service) LogoutAllDevices(ctx context.Context, userID uint) error {
	s.logger.Info().Uint("user_id", userID).Msg("Logout all devices initiated")

	// TODO: If you track user sessions in database, you can blacklist all their tokens
	// For now, this would require storing session tokens in database
	// which is beyond the scope of this implementation

	// Alternative: Set a "logged_out_at" timestamp on user record
	// and reject any tokens issued before that timestamp

	return nil
}

// ValidateToken validates a JWT token and checks if it's blacklisted
func (s *Service) ValidateToken(tokenString string) (*user.User, error) {
	// First validate the token structure and signature
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is blacklisted
	isBlacklisted, err := s.tokenBlacklist.IsBlacklisted(context.Background(), tokenString)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to check token blacklist")
		// Continue anyway - don't fail auth due to blacklist check failure
	}

	if isBlacklisted {
		return nil, e.ErrTokenInvalid.WithDetail("reason", "token has been revoked")
	}

	// Get user from database
	user, err := s.repo.GetUserByID(context.Background(), claims.UserID)
	if err != nil {
		return nil, err
	}

	// Remove password from response
	user.Password = ""

	return user, nil
}

// Helper functions

// generateVerificationCode generates a random 6-digit verification code
func (s *Service) generateVerificationCode() (string, error) {
	// Generate 6-digit code
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Format as 6-digit string with leading zeros
	code := fmt.Sprintf("%06d", n.Int64())
	return code, nil
}

// generateResetToken generates a secure random token
func (s *Service) generateResetToken() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Convert to hex string
	token := fmt.Sprintf("%x", b)
	return token, nil
}