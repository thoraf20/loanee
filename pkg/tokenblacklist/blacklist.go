package tokenblacklist

import (
	"context"
	"time"
)

// Blacklist defines the interface for token blacklist operations
type Blacklist interface {
	Add(ctx context.Context, token string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	Remove(ctx context.Context, token string) error
	Cleanup(ctx context.Context) error
}