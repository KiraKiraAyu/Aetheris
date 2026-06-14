package main

import (
	"context"
	"errors"
	"log"

	"github.com/hibiken/asynq"

	"aetheris/internal/cache"
	"aetheris/internal/config"
	"aetheris/internal/database"
	"aetheris/internal/delivery"
	"aetheris/internal/jobs"
	"aetheris/internal/notification"
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

	repo := notification.NewGormRepository(db)
	dispatcher := delivery.NewConfiguredDispatcher(repo, repo)

	if cfg.QueueType == "redis" && cfg.RedisAddr != "" {
		server := asynq.NewServer(
			cache.AsynqRedisOpt(cfg),
			asynq.Config{
				Concurrency: cfg.WorkerConcurrency,
				Queues: map[string]int{
					cfg.QueueName: 1,
				},
			},
		)
		handler := jobs.NewDeliveryHandler(repo, dispatcher, jobs.SystemClock{})
		mux := asynq.NewServeMux()
		handler.Register(mux)
		if err := server.Run(mux); err != nil {
			log.Fatalf("run worker: %v", err)
		}
	} else {
		dbWorker := jobs.NewDBWorker(repo, dispatcher, cfg.WorkerConcurrency, jobs.SystemClock{})
		log.Println("Starting database-backed queue worker...")
		if err := dbWorker.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("run db worker: %v", err)
		}
	}
}
