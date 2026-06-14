package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"aetheris/internal/notification"
)

const defaultQueueName = "notifications"

type Enqueuer struct {
	client    *asynq.Client
	queueName string
	uniqueTTL time.Duration
}

func NewEnqueuer(client *asynq.Client, queueName string, uniqueTTL time.Duration) *Enqueuer {
	if queueName == "" {
		queueName = defaultQueueName
	}
	if uniqueTTL <= 0 {
		uniqueTTL = 5 * time.Minute
	}
	return &Enqueuer{
		client:    client,
		queueName: queueName,
		uniqueTTL: uniqueTTL,
	}
}

func (e *Enqueuer) EnqueueDelivery(ctx context.Context, notification notification.Notification) error {
	if e.client == nil {
		return fmt.Errorf("asynq enqueuer: client is required")
	}
	task := NewDeliveryTask(DeliveryPayload{NotificationID: notification.ID})
	_, err := e.client.EnqueueContext(
		ctx,
		task,
		asynq.Queue(e.queueName),
		asynq.TaskID("deliver:"+notification.ID),
		asynq.Unique(e.uniqueTTL),
		asynq.MaxRetry(5),
	)
	if errors.Is(err, asynq.ErrDuplicateTask) || errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}
