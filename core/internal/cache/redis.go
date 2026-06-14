package cache

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"aetheris/internal/config"
)

func NewRedisClient(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

func AsynqRedisOpt(cfg config.Config) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
}
