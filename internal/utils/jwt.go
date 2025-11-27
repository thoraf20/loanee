package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thoraf20/loanee/config"
	e "github.com/thoraf20/loanee/pkg/error"
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles JWT operations with injected config
type Manager struct {
	config *config.Config
}

// NewManager creates a new JWT manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (m *Manager) GenerateTokenPair(userID uuid.UUID, email, name, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = m.GenerateAccessToken(userID, email, name, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.GenerateRefreshToken(userID, email, name, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken creates a new access token
func (m *Manager) GenerateAccessToken(userID uuid.UUID, email, name, role string) (string, error) {
	if m.config.JWT.Secret == "" {
		return "", e.NewInternalError("JWT secret is not configured", nil)
	}

	now := time.Now()
	expiresAt := now.Add(m.config.JWT.AccessTokenExpiry)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.App.Name,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        uuid.New().String(), // Unique token ID
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.JWT.Secret))
	if err != nil {
		return "", e.NewInternalError("failed to sign token", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new refresh token with longer expiry
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, email, name, role string) (string, error) {
	if m.config.JWT.Secret == "" {
		return "", e.NewInternalError("JWT secret is not configured", nil)
	}

	now := time.Now()
	expiresAt := now.Add(m.config.JWT.RefreshTokenExpiry)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.App.Name,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.JWT.Secret))
	if err != nil {
		return "", e.NewInternalError("failed to sign refresh token", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	if m.config.JWT.Secret == "" {
		return nil, e.NewInternalError("JWT secret is not configured", nil)
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, e.ErrTokenInvalid.WithDetail("signing_method", token.Header["alg"])
			}
			return []byte(m.config.JWT.Secret), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, e.ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, e.ErrTokenInvalid.WithDetail("reason", "malformed token")
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, e.ErrTokenInvalid.WithDetail("reason", "invalid signature")
		}
		return nil, e.ErrTokenInvalid.WithError(err)
	}

	if !token.Valid {
		return nil, e.ErrTokenInvalid.WithDetail("reason", "token validation failed")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, e.ErrTokenInvalid.WithDetail("reason", "invalid claims structure")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (m *Manager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token with same user info
	return m.GenerateAccessToken(claims.UserID, claims.Email, claims.Name, claims.Role)
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", e.ErrTokenMissing
	}

	// Expected format: "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", e.ErrTokenInvalid.WithDetail("reason", "invalid authorization header format")
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", e.ErrTokenInvalid.WithDetail("reason", "missing Bearer prefix")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", e.ErrTokenMissing
	}

	return token, nil
}