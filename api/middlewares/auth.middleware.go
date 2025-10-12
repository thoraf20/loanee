package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/thoraf20/loanee/internal/utils"
)

// AuthMiddleware checks for valid JWT and injects user ID into request context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn().Msg("Missing Authorization header")
			utils.Error(w, http.StatusUnauthorized, "Unauthorized", "")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn().Msg("Invalid Authorization header format")
			utils.Error(w, http.StatusUnauthorized, "Invalid Authorization header format", "")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid or expired token")
			utils.Error(w, http.StatusUnauthorized, "Invalid or expired token", "")
			return
		}

		// Add the user ID to the request context
		ctx := utils.SetUserIDInContext(r.Context(), claims.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}