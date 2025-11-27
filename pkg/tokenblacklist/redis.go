package tokenblacklist

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RedisBlacklist struct {
	client *redis.Client
	logger zerolog.Logger
	prefix string
}

// NewRedisBlacklist creates a new Redis-based token blacklist
func NewRedisBlacklist(client *redis.Client, logger zerolog.Logger) Blacklist {
	return &RedisBlacklist{
		client: client,
		logger: logger,
		prefix: "token_blacklist:",
	}
}

func (r *RedisBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
	key := r.prefix + token
	
	// Calculate TTL
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Store token in Redis with expiration
	err := r.client.Set(ctx, key, "blacklisted", ttl).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("token", token[:20]+"...").Msg("Failed to add token to blacklist")
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	r.logger.Info().
		Str("token_prefix", token[:20]+"...").
		Time("expires_at", expiresAt).
		Msg("Token added to blacklist")

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := r.prefix + token

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check token blacklist")
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists > 0, nil
}

// removes a token from the blacklist
func (r *RedisBlacklist) Remove(ctx context.Context, token string) error {
	key := r.prefix + token

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to remove token from blacklist")
		return fmt.Errorf("failed to remove token: %w", err)
	}

	return nil
}

// Cleanup is not needed for Redis as it handles TTL automatically
func (r *RedisBlacklist) Cleanup(ctx context.Context) error {
	return nil
}