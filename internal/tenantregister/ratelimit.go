package tenantregister

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	prefix string
	limit  int
}

func NewRateLimiter(client *redis.Client, prefix string, limit int) *RateLimiter {
	if limit <= 0 {
		limit = 5
	}
	if prefix == "" {
		prefix = "monti_jarvis:"
	}
	return &RateLimiter{redis: client, prefix: prefix, limit: limit}
}

func (r *RateLimiter) Allow(ctx context.Context, clientIP string) (bool, error) {
	if r == nil || r.redis == nil || clientIP == "" {
		return true, nil
	}
	key := r.prefix + "register:ip:" + clientIP
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

func (r *RateLimiter) Key(clientIP string) string {
	return fmt.Sprintf("%sregister:ip:%s", r.prefix, clientIP)
}