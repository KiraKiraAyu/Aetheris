package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"aetheris/internal/notification"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type DeliveryResult struct {
	ProviderMessageID string
}

type Dispatcher interface {
	Deliver(context.Context, notification.Notification) (DeliveryResult, error)
}

type DeliveryHandler struct {
	repo       notification.DeliveryRepository
	dispatcher Dispatcher
	clock      Clock
}

func NewDeliveryHandler(repo notification.DeliveryRepository, dispatcher Dispatcher, clock Clock) *DeliveryHandler {
	if clock == nil {
		clock = SystemClock{}
	}
	return &DeliveryHandler{
		repo:       repo,
		dispatcher: dispatcher,
		clock:      clock,
	}
}

func (h *DeliveryHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeDeliverNotification, h.ProcessTask)
}

func (h *DeliveryHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != TypeDeliverNotification {
		return fmt.Errorf("unsupported task type %q", task.Type())
	}

	var payload DeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode delivery task: %w", err)
	}
	if payload.NotificationID == "" {
		return fmt.Errorf("decode delivery task: notification_id is required")
	}
	if h.repo == nil {
		return fmt.Errorf("delivery handler: repository is required")
	}
	if h.dispatcher == nil {
		return fmt.Errorf("delivery handler: dispatcher is required")
	}

	notificationRecord, err := h.repo.GetByID(ctx, payload.NotificationID)
	if err != nil {
		return err
	}

	attempt, err := h.beginAttempt(ctx, notificationRecord)
	if err != nil {
		return err
	}

	result, err := h.dispatcher.Deliver(ctx, notificationRecord)
	now := h.clock.Now().UTC()
	if err != nil {
		markErr := h.repo.MarkDeliveryResult(ctx, notificationRecord.ID, notification.DeliveryUpdate{
			Status:    notification.StatusFailed,
			LastError: err.Error(),
		})
		finishErr := h.finishAttempt(ctx, attempt, notification.AttemptStatusFailed, "", err.Error(), now)
		if markErr != nil || finishErr != nil {
			return errors.Join(err, markErr, finishErr)
		}
		if IsPermanent(err) {
			return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
		}
		return err
	}

	markErr := h.repo.MarkDeliveryResult(ctx, notificationRecord.ID, notification.DeliveryUpdate{
		Status:            notification.StatusDelivered,
		ProviderMessageID: result.ProviderMessageID,
		DeliveredAt:       &now,
	})
	finishErr := h.finishAttempt(ctx, attempt, notification.AttemptStatusDelivered, result.ProviderMessageID, "", now)
	return errors.Join(markErr, finishErr)
}

func (h *DeliveryHandler) beginAttempt(ctx context.Context, record notification.Notification) (notification.DeliveryAttempt, error) {
	count, err := h.repo.CountDeliveryAttempts(ctx, record.ID)
	if err != nil {
		return notification.DeliveryAttempt{}, err
	}
	attempt := notification.DeliveryAttempt{
		NotificationID: record.ID,
		TenantID:       record.TenantID,
		Channel:        record.Channel,
		Attempt:        count + 1,
		Status:         notification.AttemptStatusRunning,
		StartedAt:      h.clock.Now().UTC(),
	}
	if err := h.repo.CreateDeliveryAttempt(ctx, &attempt); err != nil {
		return notification.DeliveryAttempt{}, err
	}
	return attempt, nil
}

func (h *DeliveryHandler) finishAttempt(ctx context.Context, attempt notification.DeliveryAttempt, status notification.AttemptStatus, providerMessageID string, lastError string, finishedAt time.Time) error {
	return h.repo.FinishDeliveryAttempt(ctx, attempt.ID, notification.DeliveryAttemptUpdate{
		Status:            status,
		ProviderMessageID: providerMessageID,
		LastError:         lastError,
		FinishedAt:        &finishedAt,
		DurationMS:        finishedAt.Sub(attempt.StartedAt).Milliseconds(),
	})
}
