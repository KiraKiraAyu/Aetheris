package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"aetheris/internal/notification"
)

type fakePollingRepository struct {
	mu                 sync.Mutex
	notifs             []notification.Notification
	attempts           map[string]int
	createAttemptCalls int
	statusUpdateCalls  map[string][]notification.Status
}

func (r *fakePollingRepository) GetByID(ctx context.Context, id string) (notification.Notification, error) {
	return notification.Notification{}, nil
}

func (r *fakePollingRepository) FinishDeliveryAttempt(ctx context.Context, id string, update notification.DeliveryAttemptUpdate) error {
	return nil
}

func (r *fakePollingRepository) GetChannelConfig(ctx context.Context, tenantID string, channel notification.Channel) (notification.ChannelConfig, error) {
	return notification.ChannelConfig{}, nil
}

func (r *fakePollingRepository) PollQueued(ctx context.Context, limit int, failedSince time.Time, runningSince time.Time) ([]notification.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var eligible []notification.Notification
	for _, n := range r.notifs {
		if n.Status == notification.StatusQueued ||
			(n.Status == notification.StatusFailed && n.UpdatedAt.Before(failedSince)) ||
			(n.Status == notification.StatusRunning && n.UpdatedAt.Before(runningSince)) {
			eligible = append(eligible, n)
		}
	}
	return eligible, nil
}

func (r *fakePollingRepository) CountDeliveryAttempts(ctx context.Context, notificationID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.attempts[notificationID], nil
}

func (r *fakePollingRepository) CreateDeliveryAttempt(ctx context.Context, attempt *notification.DeliveryAttempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.createAttemptCalls++
	return nil
}

func (r *fakePollingRepository) MarkDeliveryResult(ctx context.Context, id string, result notification.DeliveryUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusUpdateCalls[id] = append(r.statusUpdateCalls[id], result.Status)
	for i, n := range r.notifs {
		if n.ID == id {
			r.notifs[i].Status = result.Status
			r.notifs[i].UpdatedAt = time.Now().UTC()
		}
	}
	return nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestDBWorkerProcessesQueuedNotification(t *testing.T) {
	repo := &fakePollingRepository{
		notifs: []notification.Notification{
			{
				ID:        "notif_ok",
				Status:    notification.StatusQueued,
				UpdatedAt: time.Now().UTC().Add(-10 * time.Second),
			},
		},
		attempts:          map[string]int{"notif_ok": 0},
		statusUpdateCalls: make(map[string][]notification.Status),
	}
	dispatcher := &fakeDispatcher{
		result: DeliveryResult{ProviderMessageID: "p123"},
	}

	worker := NewDBWorker(repo, dispatcher, 2, fixedClock{now: time.Now().UTC()})
	worker.SetPollInterval(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	// Wait for processing to complete
	time.Sleep(100 * time.Millisecond)
	cancel()

	repo.mu.Lock()
	defer repo.mu.Unlock()

	updates := repo.statusUpdateCalls["notif_ok"]
	if len(updates) < 2 {
		t.Fatalf("expected at least 2 status updates, got: %v", updates)
	}

	if updates[0] != notification.StatusRunning {
		t.Errorf("expected first update to be 'running', got %s", updates[0])
	}
	if updates[len(updates)-1] != notification.StatusDelivered {
		t.Errorf("expected final update to be 'delivered', got %s", updates[len(updates)-1])
	}
}

func TestDBWorkerSkipsFailedTaskIfAttemptsExceeded(t *testing.T) {
	repo := &fakePollingRepository{
		notifs: []notification.Notification{
			{
				ID:        "notif_failed_max",
				Status:    notification.StatusFailed,
				UpdatedAt: time.Now().UTC().Add(-2 * time.Minute),
			},
		},
		attempts:          map[string]int{"notif_failed_max": 5}, // 5 attempts
		statusUpdateCalls: make(map[string][]notification.Status),
	}
	dispatcher := &fakeDispatcher{}

	worker := NewDBWorker(repo, dispatcher, 2, fixedClock{now: time.Now().UTC()})
	worker.SetPollInterval(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	repo.mu.Lock()
	defer repo.mu.Unlock()

	updates := repo.statusUpdateCalls["notif_failed_max"]
	if len(updates) > 0 {
		t.Errorf("expected 0 status updates for exceeded attempts, got: %v", updates)
	}
}

func TestDBWorkerEnforcesBackoff(t *testing.T) {
	repo := &fakePollingRepository{
		notifs: []notification.Notification{
			{
				ID:        "notif_backoff",
				Status:    notification.StatusFailed,
				UpdatedAt: time.Now().UTC().Add(-10 * time.Second), // updated 10s ago, backoff for 1st retry is 30s
			},
		},
		attempts:          map[string]int{"notif_backoff": 1},
		statusUpdateCalls: make(map[string][]notification.Status),
	}
	dispatcher := &fakeDispatcher{}

	worker := NewDBWorker(repo, dispatcher, 2, fixedClock{now: time.Now().UTC()})
	worker.SetPollInterval(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	repo.mu.Lock()
	defer repo.mu.Unlock()

	updates := repo.statusUpdateCalls["notif_backoff"]
	if len(updates) > 0 {
		t.Errorf("expected 0 status updates due to backoff, got: %v", updates)
	}
}
