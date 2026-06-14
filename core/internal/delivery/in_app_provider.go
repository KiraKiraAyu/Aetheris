package delivery

import (
	"context"
	"fmt"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type InAppProvider struct {
	store notification.InAppStore
}

func NewInAppProvider(store notification.InAppStore) *InAppProvider {
	return &InAppProvider{store: store}
}

func (p *InAppProvider) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if p.store == nil {
		return jobs.DeliveryResult{}, fmt.Errorf("in-app delivery: store is required")
	}
	message := notification.InAppMessage{
		NotificationID: record.ID,
		TenantID:       record.TenantID,
		UserID:         record.Recipient,
		Title:          record.Title,
		Body:           record.Body,
		Metadata:       record.Metadata.Clone(),
		CreatedAt:      time.Now().UTC(),
	}
	if err := p.store.CreateInAppMessage(ctx, &message); err != nil {
		return jobs.DeliveryResult{}, err
	}
	messageID := message.ID
	if messageID == "" {
		messageID = record.ID
	}
	return jobs.DeliveryResult{ProviderMessageID: "inapp:" + messageID}, nil
}
