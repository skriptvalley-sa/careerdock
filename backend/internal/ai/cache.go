package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache TTLs per operation type.
const (
	CacheTTLResumeParse = 24 * time.Hour     // resume parsing results
	CacheTTLATSGeneral  = 24 * time.Hour     // general ATS scores
	CacheTTLATSCompany  = 7 * 24 * time.Hour // company ATS scores (stable)
	CacheTTLATSJob      = 24 * time.Hour     // job ATS scores
)

// ResultCache provides Redis-backed caching for AI operation results.
// Cache keys are SHA256 hashes of the operation name + input data.
type ResultCache struct {
	redis *redis.Client
}

// NewResultCache creates a new AI result cache.
func NewResultCache(redisClient *redis.Client) *ResultCache {
	return &ResultCache{redis: redisClient}
}

// Get retrieves a cached result. Returns nil if not found or on error.
func (c *ResultCache) Get(ctx context.Context, operation string, input string) ([]byte, error) {
	key := c.buildKey(operation, input)

	val, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		slog.Warn("ai cache get error", "key", key, "error", err)
		return nil, nil // degrade gracefully — treat as miss
	}

	slog.Debug("ai cache hit", "operation", operation, "key", key)
	return val, nil
}

// Set stores a result in the cache with the given TTL.
func (c *ResultCache) Set(ctx context.Context, operation string, input string, result []byte, ttl time.Duration) error {
	key := c.buildKey(operation, input)

	if err := c.redis.Set(ctx, key, result, ttl).Err(); err != nil {
		slog.Warn("ai cache set error", "key", key, "error", err)
		return err // non-fatal — the caller can continue without caching
	}

	slog.Debug("ai cache set", "operation", operation, "key", key, "ttl", ttl)
	return nil
}

// buildKey creates a deterministic cache key from the operation and input.
func (c *ResultCache) buildKey(operation, input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("ai:cache:%s:%s", operation, hex.EncodeToString(hash[:]))
}

// CacheKeyForResumeParse generates the cache input string for resume parsing.
func CacheKeyForResumeParse(resumeText string) string {
	return resumeText
}

// CacheKeyForATSGeneral generates the cache input string for general ATS scoring.
func CacheKeyForATSGeneral(resumeText string) string {
	return resumeText
}
