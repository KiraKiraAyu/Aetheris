package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"aetheris/internal/cache"
	"aetheris/internal/config"
	"aetheris/internal/database"
	"aetheris/internal/httpapi"
	"aetheris/internal/jobs"
	"aetheris/internal/notification"
	"aetheris/internal/ratelimit"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	var queue notification.DeliveryQueue
	var closeAsynq func() error

	if cfg.QueueType == "redis" && cfg.RedisAddr != "" {
		asynqClient := asynq.NewClient(cache.AsynqRedisOpt(cfg))
		closeAsynq = asynqClient.Close
		queue = jobs.NewEnqueuer(asynqClient, cfg.QueueName, cfg.UniqueTTL)
	} else {
		queue = notification.DBQueue{}
	}
	if closeAsynq != nil {
		defer closeAsynq()
	}

	repo := notification.NewGormRepository(db)
	service := notification.NewService(repo, queue, notification.SystemClock{})

	router := gin.Default()
	options := httpapi.Options{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		MaxBodyBytes:   cfg.RequestMaxBytes,
	}
	if len(cfg.APIKeys) > 0 {
		options.Authenticator = httpapi.NewStaticAPIKeyAuthenticator(cfg.APIKeys)
	}
	if cfg.RateLimitEnabled {
		if cfg.RedisAddr == "" {
			log.Println("WARNING: Rate limiting is enabled but REDIS_ADDR is empty. Rate limiting will be disabled.")
		} else {
			redisClient := cache.NewRedisClient(cfg)
			defer redisClient.Close()
			options.RateLimiter = ratelimit.NewRedisFixedWindowLimiter(redisClient, cfg.RateLimitPerMinute, time.Minute)
		}
	}
	httpapi.RegisterRoutesWithOptions(router, service, options)

	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run api: %v", err)
	}
}
