package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeDeliverNotification = "notification:deliver"

type DeliveryPayload struct {
	NotificationID string `json:"notification_id"`
}

func NewDeliveryTask(payload DeliveryPayload) *asynq.Task {
	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return asynq.NewTask(TypeDeliverNotification, encoded)
}
