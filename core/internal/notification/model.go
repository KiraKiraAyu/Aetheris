package notification

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelWebhook  Channel = "webhook"
	ChannelInApp    Channel = "in_app"
	ChannelTelegram Channel = "telegram"
	ChannelSlack    Channel = "slack"
	ChannelDiscord  Channel = "discord"
	ChannelFeishu   Channel = "feishu"
	ChannelDingTalk Channel = "dingtalk"
	ChannelWeCom    Channel = "wecom"
)

func (c Channel) Valid() bool {
	switch c {
	case ChannelEmail,
		ChannelSMS,
		ChannelWebhook,
		ChannelInApp,
		ChannelTelegram,
		ChannelSlack,
		ChannelDiscord,
		ChannelFeishu,
		ChannelDingTalk,
		ChannelWeCom:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type AttemptStatus string

const (
	AttemptStatusRunning   AttemptStatus = "running"
	AttemptStatusDelivered AttemptStatus = "delivered"
	AttemptStatusFailed    AttemptStatus = "failed"
)

type Metadata map[string]string

func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (m *Metadata) Scan(value any) error {
	if value == nil {
		*m = Metadata{}
		return nil
	}

	var payload []byte
	switch typed := value.(type) {
	case []byte:
		payload = typed
	case string:
		payload = []byte(typed)
	default:
		return fmt.Errorf("scan notification metadata: unsupported type %T", value)
	}

	if len(payload) == 0 {
		*m = Metadata{}
		return nil
	}
	return json.Unmarshal(payload, m)
}

func (m Metadata) Clone() Metadata {
	if len(m) == 0 {
		return Metadata{}
	}
	clone := make(Metadata, len(m))
	for key, value := range m {
		clone[key] = value
	}
	return clone
}

type Notification struct {
	ID                string     `json:"id" gorm:"type:uuid;primaryKey"`
	TenantID          string     `json:"tenant_id" gorm:"size:128;not null;index;index:idx_notifications_aggregate,priority:1;uniqueIndex:idx_notifications_idempotency,priority:1"`
	Recipient         string     `json:"recipient" gorm:"size:256;not null;index;index:idx_notifications_aggregate,priority:2"`
	Channel           Channel    `json:"channel" gorm:"size:32;not null;index:idx_notifications_aggregate,priority:3"`
	TemplateKey       string     `json:"template_key,omitempty" gorm:"size:128;index"`
	Title             string     `json:"title" gorm:"size:256;not null"`
	Body              string     `json:"body" gorm:"type:text;not null"`
	GroupKey          string     `json:"group_key,omitempty" gorm:"size:256;index:idx_notifications_aggregate,priority:4"`
	Status            Status     `json:"status" gorm:"size:32;not null;index;index:idx_notifications_aggregate,priority:5"`
	IdempotencyKey    string     `json:"idempotency_key,omitempty" gorm:"size:256;uniqueIndex:idx_notifications_idempotency,priority:2,where:idempotency_key <> ''"`
	AggregateCount    int        `json:"aggregate_count" gorm:"not null;default:1"`
	Metadata          Metadata   `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	ProviderMessageID string     `json:"provider_message_id,omitempty" gorm:"size:256"`
	LastError         string     `json:"last_error,omitempty" gorm:"type:text"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type InAppMessage struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey"`
	NotificationID string     `json:"notification_id" gorm:"type:uuid;not null;index"`
	TenantID       string     `json:"tenant_id" gorm:"size:128;not null;index"`
	UserID         string     `json:"user_id" gorm:"size:256;not null;index"`
	Title          string     `json:"title" gorm:"size:256;not null"`
	Body           string     `json:"body" gorm:"type:text;not null"`
	Metadata       Metadata   `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	ReadAt         *time.Time `json:"read_at,omitempty" gorm:"index"`
	CreatedAt      time.Time  `json:"created_at"`
}

type DeliveryAttempt struct {
	ID                string        `json:"id" gorm:"type:uuid;primaryKey"`
	NotificationID    string        `json:"notification_id" gorm:"type:uuid;not null;index"`
	TenantID          string        `json:"tenant_id" gorm:"size:128;not null;index"`
	Channel           Channel       `json:"channel" gorm:"size:32;not null;index"`
	Attempt           int           `json:"attempt" gorm:"not null"`
	Status            AttemptStatus `json:"status" gorm:"size:32;not null;index"`
	ProviderMessageID string        `json:"provider_message_id,omitempty" gorm:"size:256"`
	LastError         string        `json:"last_error,omitempty" gorm:"type:text"`
	StartedAt         time.Time     `json:"started_at"`
	FinishedAt        *time.Time    `json:"finished_at,omitempty"`
	DurationMS        int64         `json:"duration_ms"`
}

type DeliveryAttemptUpdate struct {
	Status            AttemptStatus
	ProviderMessageID string
	LastError         string
	FinishedAt        *time.Time
	DurationMS        int64
}

type NotificationTemplate struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey"`
	TenantID      string    `json:"tenant_id" gorm:"size:128;not null;uniqueIndex:idx_notification_templates_lookup,priority:1"`
	Key           string    `json:"key" gorm:"size:128;not null;uniqueIndex:idx_notification_templates_lookup,priority:2"`
	Channel       Channel   `json:"channel" gorm:"size:32;not null;uniqueIndex:idx_notification_templates_lookup,priority:3"`
	TitleTemplate string    `json:"title_template" gorm:"type:text;not null"`
	BodyTemplate  string    `json:"body_template" gorm:"type:text;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ChannelConfig struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"size:128;not null;uniqueIndex:idx_channel_configs_lookup,priority:1"`
	Channel   Channel   `json:"channel" gorm:"size:32;not null;uniqueIndex:idx_channel_configs_lookup,priority:2"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:false"`
	Config    string    `json:"config" gorm:"type:text;not null;default:'{}'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	TenantID       string   `json:"tenant_id"`
	Recipient      string   `json:"recipient"`
	Channel        Channel  `json:"channel"`
	TemplateKey    string   `json:"template_key"`
	Title          string   `json:"title"`
	Body           string   `json:"body"`
	GroupKey       string   `json:"group_key"`
	IdempotencyKey string   `json:"idempotency_key"`
	Metadata       Metadata `json:"metadata"`
}

type AggregateKey struct {
	TenantID  string
	Recipient string
	Channel   Channel
	GroupKey  string
}

type DeliveryUpdate struct {
	Status            Status
	ProviderMessageID string
	LastError         string
	DeliveredAt       *time.Time
}

type NotificationQuery struct {
	TenantID  string
	Recipient string
	Channel   Channel
	Status    Status
	Limit     int
}

type InAppQuery struct {
	TenantID   string
	UserID     string
	UnreadOnly bool
	Limit      int
}

type TemplateQuery struct {
	TenantID string
	Channel  Channel
	Key      string
	Limit    int
}
