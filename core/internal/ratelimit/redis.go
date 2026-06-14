package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisFixedWindowLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRedisFixedWindowLimiter(client *redis.Client, limit int, window time.Duration) *RedisFixedWindowLimiter {
	if window <= 0 {
		window = time.Minute
	}
	return &RedisFixedWindowLimiter{client: client, limit: limit, window: window}
}

func (l *RedisFixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if l == nil || l.client == nil || l.limit <= 0 {
		return true, nil
	}
	redisKey := "rate_limit:" + key
	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, redisKey, l.window).Err(); err != nil {
			return false, err
		}
	}
	return count <= int64(l.limit), nil
}
