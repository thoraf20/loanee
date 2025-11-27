package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserIDFromGin safely extracts the authenticated user ID from gin context.
func UserIDFromGin(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}

// MustUserID returns the user ID or panics; intended for tests only.
func MustUserID(c *gin.Context) uuid.UUID {
	if id, ok := UserIDFromGin(c); ok {
		return id
	}
	panic(errors.New("user_id missing in context"))
}
