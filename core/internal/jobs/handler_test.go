package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"aetheris/internal/notification"
)

func TestDeliveryHandlerMarksNotificationDelivered(t *testing.T) {
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{
			ID:        "notif_1",
			TenantID:  "tenant-a",
			Recipient: "user-1",
			Channel:   notification.ChannelEmail,
			Status:    notification.StatusQueued,
		},
	}
	dispatcher := &fakeDispatcher{
		result: DeliveryResult{ProviderMessageID: "provider_123"},
	}
	handler := NewDeliveryHandler(repo, dispatcher, fixedJobClock{now: time.Date(2026, 6, 6, 13, 0, 0, 0, time.UTC)})

	err := handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "notif_1"}))

	if err != nil {
		t.Fatalf("ProcessTask returned error: %v", err)
	}
	if dispatcher.delivered.ID != "notif_1" {
		t.Fatalf("dispatcher received notification %q, want notif_1", dispatcher.delivered.ID)
	}
	if repo.status != notification.StatusDelivered {
		t.Fatalf("status = %q, want %q", repo.status, notification.StatusDelivered)
	}
	if repo.providerMessageID != "provider_123" {
		t.Fatalf("provider message ID = %q, want provider_123", repo.providerMessageID)
	}
	if repo.deliveredAt == nil {
		t.Fatal("delivered timestamp was not recorded")
	}
	if repo.attempt.ID == "" || repo.attempt.Status != notification.AttemptStatusDelivered {
		t.Fatalf("attempt = %#v, want delivered attempt", repo.attempt)
	}
}

func TestDeliveryHandlerMarksNotificationFailedWhenDispatcherFails(t *testing.T) {
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{
			ID:        "notif_2",
			TenantID:  "tenant-a",
			Recipient: "user-2",
			Channel:   notification.ChannelSMS,
			Status:    notification.StatusQueued,
		},
	}
	dispatcher := &fakeDispatcher{err: errors.New("provider unavailable")}
	handler := NewDeliveryHandler(repo, dispatcher, fixedJobClock{now: time.Date(2026, 6, 6, 13, 5, 0, 0, time.UTC)})

	err := handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "notif_2"}))

	if err == nil {
		t.Fatal("ProcessTask should return dispatcher error")
	}
	if repo.status != notification.StatusFailed {
		t.Fatalf("status = %q, want %q", repo.status, notification.StatusFailed)
	}
	if repo.lastError != "provider unavailable" {
		t.Fatalf("last error = %q, want provider unavailable", repo.lastError)
	}
	if repo.attempt.Status != notification.AttemptStatusFailed || repo.attempt.LastError != "provider unavailable" {
		t.Fatalf("attempt = %#v, want failed attempt", repo.attempt)
	}
}

func TestDeliveryHandlerSkipsRetryForPermanentErrors(t *testing.T) {
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{
			ID:        "notif_permanent",
			TenantID:  "tenant-a",
			Recipient: "user-1",
			Channel:   notification.ChannelWebhook,
			Status:    notification.StatusQueued,
		},
	}
	handler := NewDeliveryHandler(repo, &fakeDispatcher{err: Permanent(errors.New("bad recipient"))}, fixedJobClock{now: time.Now()})

	err := handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "notif_permanent"}))

	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, want asynq.SkipRetry", err)
	}
	if repo.status != notification.StatusFailed {
		t.Fatalf("status = %q, want failed", repo.status)
	}
	if repo.attempt.Status != notification.AttemptStatusFailed {
		t.Fatalf("attempt = %#v, want failed", repo.attempt)
	}
}

func TestDeliveryHandlerReturnsRepositoryErrors(t *testing.T) {
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{ID: "other"},
	}
	handler := NewDeliveryHandler(repo, &fakeDispatcher{}, fixedJobClock{now: time.Now()})

	err := handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "missing"}))

	if !errors.Is(err, notification.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestDeliveryHandlerReturnsDispatcherAndMarkErrors(t *testing.T) {
	dispatchErr := errors.New("provider unavailable")
	markErr := errors.New("postgres unavailable")
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{
			ID:        "notif_3",
			TenantID:  "tenant-a",
			Recipient: "user-3",
			Channel:   notification.ChannelWebhook,
			Status:    notification.StatusQueued,
		},
		markErr: markErr,
	}
	dispatcher := &fakeDispatcher{err: dispatchErr}
	handler := NewDeliveryHandler(repo, dispatcher, fixedJobClock{now: time.Now()})

	err := handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "notif_3"}))

	if !errors.Is(err, dispatchErr) || !errors.Is(err, markErr) {
		t.Fatalf("error = %v, want dispatcher and mark errors", err)
	}
}

func TestDeliveryHandlerRegistersWithAsynqServeMux(t *testing.T) {
	repo := &fakeDeliveryRepository{
		notification: notification.Notification{
			ID:        "notif_4",
			TenantID:  "tenant-a",
			Recipient: "user-4",
			Channel:   notification.ChannelInApp,
			Status:    notification.StatusQueued,
		},
	}
	handler := NewDeliveryHandler(repo, &fakeDispatcher{}, fixedJobClock{now: time.Now()})
	mux := asynq.NewServeMux()
	handler.Register(mux)

	err := mux.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{NotificationID: "notif_4"}))

	if err != nil {
		t.Fatalf("mux ProcessTask returned error: %v", err)
	}
	if repo.status != notification.StatusDelivered {
		t.Fatalf("status = %q, want delivered", repo.status)
	}
}

func TestNewDeliveryTaskEncodesPayload(t *testing.T) {
	task := NewDeliveryTask(DeliveryPayload{NotificationID: "notif_1"})

	if task.Type() != TypeDeliverNotification {
		t.Fatalf("task type = %q, want %q", task.Type(), TypeDeliverNotification)
	}
	var payload DeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if payload.NotificationID != "notif_1" {
		t.Fatalf("notification ID = %q, want notif_1", payload.NotificationID)
	}
}

func TestDeliveryHandlerRejectsBadTasks(t *testing.T) {
	handler := NewDeliveryHandler(&fakeDeliveryRepository{}, &fakeDispatcher{}, fixedJobClock{now: time.Now()})

	err := handler.ProcessTask(context.Background(), asynq.NewTask("other", nil))
	if err == nil || !strings.Contains(err.Error(), "unsupported task type") {
		t.Fatalf("error = %v, want unsupported task type", err)
	}

	err = handler.ProcessTask(context.Background(), asynq.NewTask(TypeDeliverNotification, []byte(`{`)))
	if err == nil || !strings.Contains(err.Error(), "decode delivery task") {
		t.Fatalf("error = %v, want decode error", err)
	}

	err = handler.ProcessTask(context.Background(), NewDeliveryTask(DeliveryPayload{}))
	if err == nil || !strings.Contains(err.Error(), "notification_id") {
		t.Fatalf("error = %v, want missing notification_id", err)
	}
}

func TestDeliveryHandlerRequiresDependencies(t *testing.T) {
	task := NewDeliveryTask(DeliveryPayload{NotificationID: "notif_1"})

	err := NewDeliveryHandler(nil, &fakeDispatcher{}, fixedJobClock{now: time.Now()}).ProcessTask(context.Background(), task)
	if err == nil || !strings.Contains(err.Error(), "repository") {
		t.Fatalf("error = %v, want repository dependency error", err)
	}

	err = NewDeliveryHandler(&fakeDeliveryRepository{}, nil, fixedJobClock{now: time.Now()}).ProcessTask(context.Background(), task)
	if err == nil || !strings.Contains(err.Error(), "dispatcher") {
		t.Fatalf("error = %v, want dispatcher dependency error", err)
	}
}

func TestEnqueuerRequiresClient(t *testing.T) {
	enqueuer := NewEnqueuer(nil, "", 0)

	err := enqueuer.EnqueueDelivery(context.Background(), notification.Notification{ID: "notif_1"})

	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("error = %v, want client dependency error", err)
	}
}

func TestNewEnqueuerDefaults(t *testing.T) {
	enqueuer := NewEnqueuer(nil, "", 0)

	if enqueuer.queueName != defaultQueueName {
		t.Fatalf("queue name = %q, want %q", enqueuer.queueName, defaultQueueName)
	}
	if enqueuer.uniqueTTL != 5*time.Minute {
		t.Fatalf("unique ttl = %s, want 5m", enqueuer.uniqueTTL)
	}
}

type fixedJobClock struct {
	now time.Time
}

func (c fixedJobClock) Now() time.Time {
	return c.now
}

type fakeDeliveryRepository struct {
	notification      notification.Notification
	status            notification.Status
	providerMessageID string
	lastError         string
	deliveredAt       *time.Time
	markErr           error
	attempt           notification.DeliveryAttempt
}

func (r *fakeDeliveryRepository) GetByID(_ context.Context, id string) (notification.Notification, error) {
	if r.notification.ID != id {
		return notification.Notification{}, notification.ErrNotFound
	}
	return r.notification, nil
}

func (r *fakeDeliveryRepository) MarkDeliveryResult(_ context.Context, id string, result notification.DeliveryUpdate) error {
	if r.notification.ID != id {
		return notification.ErrNotFound
	}
	r.status = result.Status
	r.providerMessageID = result.ProviderMessageID
	r.lastError = result.LastError
	r.deliveredAt = result.DeliveredAt
	return r.markErr
}

func (r *fakeDeliveryRepository) CountDeliveryAttempts(_ context.Context, notificationID string) (int, error) {
	if r.notification.ID != notificationID {
		return 0, notification.ErrNotFound
	}
	return 0, nil
}

func (r *fakeDeliveryRepository) CreateDeliveryAttempt(_ context.Context, attempt *notification.DeliveryAttempt) error {
	if attempt.ID == "" {
		attempt.ID = "attempt_1"
	}
	r.attempt = *attempt
	return nil
}

func (r *fakeDeliveryRepository) FinishDeliveryAttempt(_ context.Context, id string, update notification.DeliveryAttemptUpdate) error {
	if r.attempt.ID != id {
		return notification.ErrNotFound
	}
	r.attempt.Status = update.Status
	r.attempt.ProviderMessageID = update.ProviderMessageID
	r.attempt.LastError = update.LastError
	r.attempt.FinishedAt = update.FinishedAt
	r.attempt.DurationMS = update.DurationMS
	return nil
}

func (r *fakeDeliveryRepository) GetChannelConfig(_ context.Context, tenantID string, channel notification.Channel) (notification.ChannelConfig, error) {
	return notification.ChannelConfig{}, nil
}

type fakeDispatcher struct {
	delivered notification.Notification
	result    DeliveryResult
	err       error
}

func (d *fakeDispatcher) Deliver(_ context.Context, notification notification.Notification) (DeliveryResult, error) {
	d.delivered = notification
	return d.result, d.err
}
