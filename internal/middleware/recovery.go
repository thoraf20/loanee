// internal/middleware/recovery.go
package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/utils"
)

// Recovery recovers from panics and logs them
func Recovery(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error().
					Interface("error", err).
					Str("stack", string(debug.Stack())).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("Panic recovered")

				// Return error response
				utils.InternalServerError(c, "An unexpected error occurred", nil)
				c.Abort()
			}
		}()

		c.Next()
	}
}