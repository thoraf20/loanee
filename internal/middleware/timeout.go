// internal/middleware/timeout.go
package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoraf20/loanee/internal/utils"
)

// Timeout adds a timeout to each request
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			utils.Error(c, 504, "Request timeout", nil)
			c.Abort()
		}
	}
}