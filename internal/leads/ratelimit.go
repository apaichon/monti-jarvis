package leads

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter is a fixed-window per-IP limiter backed by Redis INCR.
// Keys look like: {prefix}{clientIP} e.g. monti_jarvis:lead:rl:1.2.3.4
type RateLimiter struct {
	redis  *redis.Client
	prefix string
	limit  int
}

// NewRateLimiter builds a limiter. limit defaults to 10 when <= 0.
// prefix should already include redis namespace, e.g. "monti_jarvis:lead:rl:".
func NewRateLimiter(client *redis.Client, prefix string, limit int) *RateLimiter {
	if limit <= 0 {
		limit = 10
	}
	if prefix == "" {
		prefix = "monti_jarvis:lead:rl:"
	}
	return &RateLimiter{redis: client, prefix: prefix, limit: limit}
}

// Allow increments the IP counter for the current hour window.
// Returns true when under limit. Redis errors fail open (true, err).
func (r *RateLimiter) Allow(ctx context.Context, clientIP string) (bool, error) {
	if r == nil || r.redis == nil || clientIP == "" {
		return true, nil
	}
	key := r.prefix + clientIP
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return true, err
	}
	if count == 1 {
		_ = r.redis.Expire(ctx, key, time.Hour).Err()
	}
	if int(count) > r.limit {
		return false, nil
	}
	return true, nil
}

// Key returns the Redis key for an IP (testing/ops).
func (r *RateLimiter) Key(clientIP string) string {
	if r == nil {
		return ""
	}
	return r.prefix + clientIP
}
