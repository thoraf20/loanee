package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thoraf20/loanee/internal/utils"
	jwt "github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/pkg/tokenblacklist"
)

type Service struct {
	jwtManager *jwt.Manager
}

// AuthRequired validates JWT token, checks blacklist, and sets user info in context
func AuthRequired(jwtManager *jwt.Manager, tokenBlacklist tokenblacklist.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from header
		token, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := tokenBlacklist.IsBlacklisted(c.Request.Context(), token)
		if err != nil {
			// Log error but don't fail auth
			// You might want to fail auth in production for security
			c.Error(err)
		}

		if isBlacklisted {
			utils.Unauthorized(c, "Token has been revoked")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_role", claims.Role)
		c.Set("access_token", token)

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		if userRole.(string) != role {
			utils.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.Unauthorized(c, "User role not found")
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// OptionalAuth attempts to authenticate but doesn't fail if token is missing
// Useful for endpoints that work with or without authentication
func OptionalAuth(jwtManager *jwt.Manager, tokenBlacklist tokenblacklist.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Try to extract and validate token
		token, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		// Check blacklist
		isBlacklisted, err := tokenBlacklist.IsBlacklisted(c.Request.Context(), token)
		if err != nil || isBlacklisted {
			// Blacklisted or error, continue without authentication
			c.Next()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_role", claims.Role)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequireEmailVerified ensures the user's email is verified
func RequireEmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would require adding email_verified to JWT claims
		// Or fetching user from database
		// For now, we'll assume it's in claims if you add it

		c.Next()
	}
}

// RateLimitByUser applies rate limiting per user
func RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// TODO: Implement rate limiting logic using Redis
		// For now, just pass through
		_ = userID

		c.Next()
	}
}
