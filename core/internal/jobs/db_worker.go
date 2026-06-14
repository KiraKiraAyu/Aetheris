package jobs

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"aetheris/internal/notification"
)

type PollingRepository interface {
	notification.DeliveryRepository
	PollQueued(ctx context.Context, limit int, failedSince time.Time, runningSince time.Time) ([]notification.Notification, error)
}

type DBWorker struct {
	repo         PollingRepository
	dispatcher   Dispatcher
	concurrency  int
	clock        Clock
	processing   sync.Map
	pollInterval time.Duration
}

func NewDBWorker(repo PollingRepository, dispatcher Dispatcher, concurrency int, clock Clock) *DBWorker {
	if clock == nil {
		clock = SystemClock{}
	}
	if concurrency <= 0 {
		concurrency = 10
	}
	return &DBWorker{
		repo:         repo,
		dispatcher:   dispatcher,
		concurrency:  concurrency,
		clock:        clock,
		pollInterval: 500 * time.Millisecond,
	}
}

func (w *DBWorker) SetPollInterval(d time.Duration) {
	w.pollInterval = d
}

func (w *DBWorker) Run(ctx context.Context) error {
	tasksChan := make(chan notification.Notification, w.concurrency)

	var wg sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case notif, ok := <-tasksChan:
					if !ok {
						return
					}
					w.process(ctx, notif)
				}
			}
		}()
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(tasksChan)
			wg.Wait()
			return ctx.Err()
		case <-ticker.C:
			now := w.clock.Now().UTC()
			// Failed tasks older than 30s can be considered for retry
			failedSince := now.Add(-30 * time.Second)
			// Stuck running tasks older than 5 minutes can be re-processed
			runningSince := now.Add(-5 * time.Minute)

			notifs, err := w.repo.PollQueued(ctx, w.concurrency*2, failedSince, runningSince)
			if err != nil {
				log.Printf("db worker: poll error: %v", err)
				continue
			}

			for _, notif := range notifs {
				if _, loaded := w.processing.LoadOrStore(notif.ID, true); loaded {
					continue
				}

				// Check retry eligibility for failed tasks
				if notif.Status == notification.StatusFailed {
					attempts, err := w.repo.CountDeliveryAttempts(ctx, notif.ID)
					if err != nil {
						w.processing.Delete(notif.ID)
						continue
					}
					if attempts >= 5 {
						// Hard limit: max 5 attempts, do not process again
						w.processing.Delete(notif.ID)
						continue
					}

					// Calculate exponential backoff: 30s * 2^(attempts-1)
					// attempt 1 failed -> backoff = 30s
					// attempt 2 failed -> backoff = 60s
					// attempt 3 failed -> backoff = 120s
					backoff := time.Duration(1<<(attempts-1)) * 30 * time.Second
					if now.Sub(notif.UpdatedAt) < backoff {
						w.processing.Delete(notif.ID)
						continue
					}
				}

				select {
				case <-ctx.Done():
					w.processing.Delete(notif.ID)
					break
				case tasksChan <- notif:
					// Handed off to worker goroutines
				default:
					// Worker pool is busy, release lock for next poll
					w.processing.Delete(notif.ID)
				}
			}
		}
	}
}

func (w *DBWorker) process(ctx context.Context, notif notification.Notification) {
	defer w.processing.Delete(notif.ID)

	// Mark status as running
	err := w.repo.MarkDeliveryResult(ctx, notif.ID, notification.DeliveryUpdate{
		Status: notification.StatusRunning,
	})
	if err != nil {
		log.Printf("db worker: mark running status failed for %s: %v", notif.ID, err)
		return
	}

	attempt, err := w.beginAttempt(ctx, notif)
	if err != nil {
		log.Printf("db worker: begin attempt failed for %s: %v", notif.ID, err)
		return
	}

	result, err := w.dispatcher.Deliver(ctx, notif)
	now := w.clock.Now().UTC()
	if err != nil {
		markErr := w.repo.MarkDeliveryResult(ctx, notif.ID, notification.DeliveryUpdate{
			Status:    notification.StatusFailed,
			LastError: err.Error(),
		})
		finishErr := w.finishAttempt(ctx, attempt, notification.AttemptStatusFailed, "", err.Error(), now)
		if markErr != nil || finishErr != nil {
			log.Printf("db worker: mark failure results failed for %s: %v", notif.ID, errors.Join(err, markErr, finishErr))
		}
		return
	}

	markErr := w.repo.MarkDeliveryResult(ctx, notif.ID, notification.DeliveryUpdate{
		Status:            notification.StatusDelivered,
		ProviderMessageID: result.ProviderMessageID,
		DeliveredAt:       &now,
	})
	finishErr := w.finishAttempt(ctx, attempt, notification.AttemptStatusDelivered, result.ProviderMessageID, "", now)
	if markErr != nil || finishErr != nil {
		log.Printf("db worker: mark success results failed for %s: %v", notif.ID, errors.Join(markErr, finishErr))
	}
}

func (w *DBWorker) beginAttempt(ctx context.Context, record notification.Notification) (notification.DeliveryAttempt, error) {
	count, err := w.repo.CountDeliveryAttempts(ctx, record.ID)
	if err != nil {
		return notification.DeliveryAttempt{}, err
	}
	attempt := notification.DeliveryAttempt{
		NotificationID: record.ID,
		TenantID:       record.TenantID,
		Channel:        record.Channel,
		Attempt:        count + 1,
		Status:         notification.AttemptStatusRunning,
		StartedAt:      w.clock.Now().UTC(),
	}
	if err := w.repo.CreateDeliveryAttempt(ctx, &attempt); err != nil {
		return notification.DeliveryAttempt{}, err
	}
	return attempt, nil
}

func (w *DBWorker) finishAttempt(ctx context.Context, attempt notification.DeliveryAttempt, status notification.AttemptStatus, providerMessageID string, lastError string, finishedAt time.Time) error {
	return w.repo.FinishDeliveryAttempt(ctx, attempt.ID, notification.DeliveryAttemptUpdate{
		Status:            status,
		ProviderMessageID: providerMessageID,
		LastError:         lastError,
		FinishedAt:        &finishedAt,
		DurationMS:        finishedAt.Sub(attempt.StartedAt).Milliseconds(),
	})
}
