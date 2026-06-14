package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Notification{}, &InAppMessage{}, &DeliveryAttempt{}, &NotificationTemplate{}, &ChannelConfig{})
}

func (r *GormRepository) FindOpenAggregate(ctx context.Context, key AggregateKey) (*Notification, error) {
	if key.GroupKey == "" {
		return nil, ErrNotFound
	}

	var notification Notification
	err := r.db.WithContext(ctx).
		Where(
			"tenant_id = ? AND recipient = ? AND channel = ? AND group_key = ? AND status = ?",
			key.TenantID,
			key.Recipient,
			key.Channel,
			key.GroupKey,
			StatusQueued,
		).
		Order("updated_at DESC").
		First(&notification).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *GormRepository) Create(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *GormRepository) FindByIdempotencyKey(ctx context.Context, tenantID string, key string) (*Notification, error) {
	if key == "" {
		return nil, ErrNotFound
	}
	var notification Notification
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND idempotency_key = ?", tenantID, key).
		First(&notification).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *GormRepository) GetTemplate(ctx context.Context, tenantID string, key string, channel Channel) (NotificationTemplate, error) {
	var template NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND key = ? AND channel = ?", tenantID, key, channel).
		First(&template).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotificationTemplate{}, ErrNotFound
	}
	return template, err
}

func (r *GormRepository) UpdateAggregate(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *GormRepository) GetByID(ctx context.Context, id string) (Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Notification{}, ErrNotFound
	}
	return notification, err
}

func (r *GormRepository) GetByTenantID(ctx context.Context, tenantID string, id string) (Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&notification).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Notification{}, ErrNotFound
	}
	return notification, err
}

func (r *GormRepository) MarkDeliveryResult(ctx context.Context, id string, result DeliveryUpdate) error {
	updates := map[string]any{
		"status":              result.Status,
		"provider_message_id": result.ProviderMessageID,
		"last_error":          result.LastError,
		"updated_at":          time.Now().UTC(),
	}
	if result.DeliveredAt != nil {
		updates["delivered_at"] = *result.DeliveredAt
	}

	tx := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ?", id).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormRepository) CountDeliveryAttempts(ctx context.Context, notificationID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&DeliveryAttempt{}).
		Where("notification_id = ?", notificationID).
		Count(&count).
		Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *GormRepository) CreateDeliveryAttempt(ctx context.Context, attempt *DeliveryAttempt) error {
	if attempt.ID == "" {
		attempt.ID = uuid.NewString()
	}
	if attempt.StartedAt.IsZero() {
		attempt.StartedAt = time.Now().UTC()
	}
	if attempt.Status == "" {
		attempt.Status = AttemptStatusRunning
	}
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *GormRepository) FinishDeliveryAttempt(ctx context.Context, id string, update DeliveryAttemptUpdate) error {
	updates := map[string]any{
		"status":              update.Status,
		"provider_message_id": update.ProviderMessageID,
		"last_error":          update.LastError,
		"duration_ms":         update.DurationMS,
	}
	if update.FinishedAt != nil {
		updates["finished_at"] = *update.FinishedAt
	}
	tx := r.db.WithContext(ctx).
		Model(&DeliveryAttempt{}).
		Where("id = ?", id).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormRepository) ListDeliveryAttempts(ctx context.Context, tenantID string, notificationID string) ([]DeliveryAttempt, error) {
	var attempts []DeliveryAttempt
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND notification_id = ?", tenantID, notificationID).
		Order("attempt ASC").
		Find(&attempts).
		Error
	return attempts, err
}

func (r *GormRepository) CreateInAppMessage(ctx context.Context, message *InAppMessage) error {
	if message.ID == "" {
		message.ID = uuid.NewString()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}
	if message.Metadata == nil {
		message.Metadata = Metadata{}
	}
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *GormRepository) List(ctx context.Context, query NotificationQuery) ([]Notification, error) {
	db := r.db.WithContext(ctx).Where("tenant_id = ?", query.TenantID)
	if query.Recipient != "" {
		db = db.Where("recipient = ?", query.Recipient)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	var notifications []Notification
	err := db.Order("created_at DESC").Limit(query.Limit).Find(&notifications).Error
	return notifications, err
}

func (r *GormRepository) ListInApp(ctx context.Context, query InAppQuery) ([]InAppMessage, error) {
	db := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", query.TenantID, query.UserID)
	if query.UnreadOnly {
		db = db.Where("read_at IS NULL")
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	var messages []InAppMessage
	err := db.Order("created_at DESC").Limit(query.Limit).Find(&messages).Error
	return messages, err
}

func (r *GormRepository) MarkInAppRead(ctx context.Context, tenantID string, messageID string, userID string, readAt time.Time) error {
	tx := r.db.WithContext(ctx).
		Model(&InAppMessage{}).
		Where("tenant_id = ? AND id = ? AND user_id = ?", tenantID, messageID, userID).
		Update("read_at", readAt)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormRepository) ListTemplates(ctx context.Context, query TemplateQuery) ([]NotificationTemplate, error) {
	db := r.db.WithContext(ctx).Where("tenant_id = ?", query.TenantID)
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Key != "" {
		db = db.Where("key = ?", query.Key)
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	var templates []NotificationTemplate
	err := db.Order("key ASC, channel ASC").Limit(query.Limit).Find(&templates).Error
	return templates, err
}

func (r *GormRepository) SaveTemplate(ctx context.Context, tpl *NotificationTemplate) error {
	if tpl.ID == "" {
		tpl.ID = uuid.NewString()
		return r.db.WithContext(ctx).Create(tpl).Error
	}
	return r.db.WithContext(ctx).Save(tpl).Error
}

func (r *GormRepository) DeleteTemplate(ctx context.Context, tenantID string, id string) error {
	tx := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&NotificationTemplate{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormRepository) GetChannelConfig(ctx context.Context, tenantID string, channel Channel) (ChannelConfig, error) {
	var cfg ChannelConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND channel = ?", tenantID, channel).
		First(&cfg).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ChannelConfig{}, ErrNotFound
	}
	return cfg, err
}

func (r *GormRepository) SaveChannelConfig(ctx context.Context, cfg *ChannelConfig) error {
	if cfg.ID == "" {
		cfg.ID = uuid.NewString()
		return r.db.WithContext(ctx).Create(cfg).Error
	}
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *GormRepository) ListChannelConfigs(ctx context.Context, tenantID string) ([]ChannelConfig, error) {
	var configs []ChannelConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&configs).
		Error
	return configs, err
}

func (r *GormRepository) PollQueued(ctx context.Context, limit int, failedSince time.Time, runningSince time.Time) ([]Notification, error) {
	var notifications []Notification
	err := r.db.WithContext(ctx).
		Where("status = ?", StatusQueued).
		Or("status = ? AND updated_at < ?", StatusFailed, failedSince).
		Or("status = ? AND updated_at < ?", StatusRunning, runningSince).
		Limit(limit).
		Find(&notifications).
		Error
	return notifications, err
}
