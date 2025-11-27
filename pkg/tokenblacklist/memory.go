package tokenblacklist

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type tokenEntry struct {
	expiresAt time.Time
}

type MemoryBlacklist struct {
	tokens map[string]tokenEntry
	mu     sync.RWMutex
	logger zerolog.Logger
}

// NewMemoryBlacklist creates a new in-memory token blacklist
func NewMemoryBlacklist(logger zerolog.Logger) Blacklist {
	mb := &MemoryBlacklist{
		tokens: make(map[string]tokenEntry),
		logger: logger,
	}

	// Start cleanup goroutine
	go mb.periodicCleanup()

	return mb
}

// Add adds a token to the blacklist
func (m *MemoryBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Don't add already expired tokens
	if time.Now().After(expiresAt) {
		return nil
	}

	m.tokens[token] = tokenEntry{
		expiresAt: expiresAt,
	}

	m.logger.Info().
		Str("token_prefix", token[:20]+"...").
		Time("expires_at", expiresAt).
		Int("total_blacklisted", len(m.tokens)).
		Msg("Token added to blacklist")

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (m *MemoryBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.tokens[token]
	if !exists {
		return false, nil
	}

	// Check if token has expired
	if time.Now().After(entry.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Remove removes a token from the blacklist
func (m *MemoryBlacklist) Remove(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, token)
	return nil
}

// Cleanup removes expired tokens
func (m *MemoryBlacklist) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removedCount := 0

	for token, entry := range m.tokens {
		if now.After(entry.expiresAt) {
			delete(m.tokens, token)
			removedCount++
		}
	}

	if removedCount > 0 {
		m.logger.Info().
			Int("removed_count", removedCount).
			Int("remaining", len(m.tokens)).
			Msg("Cleaned up expired tokens")
	}

	return nil
}

// periodicCleanup runs cleanup every 5 minutes
func (m *MemoryBlacklist) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.Cleanup(context.Background()); err != nil {
			m.logger.Error().Err(err).Msg("Failed to cleanup expired tokens")
		}
	}
}