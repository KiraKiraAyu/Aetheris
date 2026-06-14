package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type Repository interface {
	FindOpenAggregate(context.Context, AggregateKey) (*Notification, error)
	FindByIdempotencyKey(context.Context, string, string) (*Notification, error)
	GetByTenantID(context.Context, string, string) (Notification, error)
	Create(context.Context, *Notification) error
	UpdateAggregate(context.Context, *Notification) error
	List(context.Context, NotificationQuery) ([]Notification, error)
	ListDeliveryAttempts(context.Context, string, string) ([]DeliveryAttempt, error)
	ListInApp(context.Context, InAppQuery) ([]InAppMessage, error)
	MarkInAppRead(context.Context, string, string, string, time.Time) error
	ListTemplates(context.Context, TemplateQuery) ([]NotificationTemplate, error)
	SaveTemplate(context.Context, *NotificationTemplate) error
	DeleteTemplate(context.Context, string, string) error
	GetChannelConfig(ctx context.Context, tenantID string, channel Channel) (ChannelConfig, error)
	SaveChannelConfig(ctx context.Context, config *ChannelConfig) error
	ListChannelConfigs(ctx context.Context, tenantID string) ([]ChannelConfig, error)
}

type DeliveryRepository interface {
	GetByID(context.Context, string) (Notification, error)
	MarkDeliveryResult(context.Context, string, DeliveryUpdate) error
	CountDeliveryAttempts(context.Context, string) (int, error)
	CreateDeliveryAttempt(context.Context, *DeliveryAttempt) error
	FinishDeliveryAttempt(context.Context, string, DeliveryAttemptUpdate) error
	GetChannelConfig(ctx context.Context, tenantID string, channel Channel) (ChannelConfig, error)
}

type InAppStore interface {
	CreateInAppMessage(context.Context, *InAppMessage) error
}

type DeliveryQueue interface {
	EnqueueDelivery(context.Context, Notification) error
}

type DBQueue struct{}

func (DBQueue) EnqueueDelivery(ctx context.Context, notif Notification) error {
	// Already persisted in GORM with status StatusQueued.
	// The database worker will poll and process it.
	return nil
}

type TemplateStore interface {
	GetTemplate(context.Context, string, string, Channel) (NotificationTemplate, error)
}

type Service struct {
	repo      Repository
	queue     DeliveryQueue
	clock     Clock
	templates TemplateStore
}

type ServiceOptions struct {
	Templates TemplateStore
}

func NewService(repo Repository, queue DeliveryQueue, clock Clock) *Service {
	return NewServiceWithOptions(repo, queue, clock, ServiceOptions{})
}

func NewServiceWithOptions(repo Repository, queue DeliveryQueue, clock Clock, options ServiceOptions) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if options.Templates == nil {
		if templates, ok := repo.(TemplateStore); ok {
			options.Templates = templates
		}
	}
	return &Service{
		repo:      repo,
		queue:     queue,
		clock:     clock,
		templates: options.Templates,
	}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (Notification, error) {
	request = normalizeCreateRequest(request)
	if err := s.renderTemplate(ctx, &request); err != nil {
		return Notification{}, err
	}
	if request.Recipient == "" {
		if s.repo != nil {
			cfg, err := s.repo.GetChannelConfig(ctx, request.TenantID, request.Channel)
			if err == nil {
				var configMap map[string]interface{}
				if err := json.Unmarshal([]byte(cfg.Config), &configMap); err == nil {
					if val, ok := configMap["default_recipient"].(string); ok {
						request.Recipient = strings.TrimSpace(val)
					}
				}
			}
		}
	}
	if err := validateCreateRequest(request); err != nil {
		return Notification{}, err
	}
	if s.repo == nil {
		return Notification{}, fmt.Errorf("notification service: repository is required")
	}
	if s.queue == nil {
		return Notification{}, fmt.Errorf("notification service: delivery queue is required")
	}
	if request.IdempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, request.TenantID, request.IdempotencyKey)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return Notification{}, err
		}
		if existing != nil {
			return *existing, nil
		}
	}

	now := s.clock.Now().UTC()
	if request.GroupKey != "" {
		existing, err := s.repo.FindOpenAggregate(ctx, AggregateKey{
			TenantID:  request.TenantID,
			Recipient: request.Recipient,
			Channel:   request.Channel,
			GroupKey:  request.GroupKey,
		})
		if err != nil && !errors.Is(err, ErrNotFound) {
			return Notification{}, err
		}
		if existing != nil {
			aggregated := mergeIntoAggregate(*existing, request, now)
			if err := s.repo.UpdateAggregate(ctx, &aggregated); err != nil {
				return Notification{}, err
			}
			if err := s.queue.EnqueueDelivery(ctx, aggregated); err != nil {
				return Notification{}, err
			}
			return aggregated, nil
		}
	}

	notification := Notification{
		ID:             uuid.NewString(),
		TenantID:       request.TenantID,
		Recipient:      request.Recipient,
		Channel:        request.Channel,
		TemplateKey:    request.TemplateKey,
		Title:          request.Title,
		Body:           request.Body,
		GroupKey:       request.GroupKey,
		IdempotencyKey: request.IdempotencyKey,
		Status:         StatusQueued,
		AggregateCount: 1,
		Metadata:       request.Metadata.Clone(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.Create(ctx, &notification); err != nil {
		return Notification{}, err
	}
	if err := s.queue.EnqueueDelivery(ctx, notification); err != nil {
		return Notification{}, err
	}
	return notification, nil
}

func (s *Service) List(ctx context.Context, query NotificationQuery) ([]Notification, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("notification service: repository is required")
	}
	query.TenantID = strings.TrimSpace(query.TenantID)
	query.Recipient = strings.TrimSpace(query.Recipient)
	if query.TenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	return s.repo.List(ctx, query)
}

func (s *Service) Get(ctx context.Context, tenantID string, id string) (Notification, error) {
	if s.repo == nil {
		return Notification{}, fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	id = strings.TrimSpace(id)
	if tenantID == "" || id == "" {
		return Notification{}, fmt.Errorf("%w: tenant_id and id are required", ErrInvalidRequest)
	}
	return s.repo.GetByTenantID(ctx, tenantID, id)
}

func (s *Service) ListDeliveryAttempts(ctx context.Context, tenantID string, notificationID string) ([]DeliveryAttempt, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	notificationID = strings.TrimSpace(notificationID)
	if tenantID == "" || notificationID == "" {
		return nil, fmt.Errorf("%w: tenant_id and notification_id are required", ErrInvalidRequest)
	}
	return s.repo.ListDeliveryAttempts(ctx, tenantID, notificationID)
}

func (s *Service) ListInApp(ctx context.Context, query InAppQuery) ([]InAppMessage, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("notification service: repository is required")
	}
	query.TenantID = strings.TrimSpace(query.TenantID)
	query.UserID = strings.TrimSpace(query.UserID)
	if query.TenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}
	if query.UserID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidRequest)
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	return s.repo.ListInApp(ctx, query)
}

func (s *Service) MarkInAppRead(ctx context.Context, tenantID string, messageID string, userID string) error {
	if s.repo == nil {
		return fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	messageID = strings.TrimSpace(messageID)
	userID = strings.TrimSpace(userID)
	if tenantID == "" || messageID == "" || userID == "" {
		return fmt.Errorf("%w: tenant_id, message_id and user_id are required", ErrInvalidRequest)
	}
	return s.repo.MarkInAppRead(ctx, tenantID, messageID, userID, s.clock.Now().UTC())
}

func (s *Service) ListTemplates(ctx context.Context, query TemplateQuery) ([]NotificationTemplate, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("notification service: repository is required")
	}
	query.TenantID = strings.TrimSpace(query.TenantID)
	query.Key = strings.TrimSpace(query.Key)
	if query.TenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}
	return s.repo.ListTemplates(ctx, query)
}

func (s *Service) SaveTemplate(ctx context.Context, tpl NotificationTemplate) (NotificationTemplate, error) {
	if s.repo == nil {
		return NotificationTemplate{}, fmt.Errorf("notification service: repository is required")
	}
	tpl.TenantID = strings.TrimSpace(tpl.TenantID)
	tpl.Key = strings.TrimSpace(tpl.Key)
	tpl.TitleTemplate = strings.TrimSpace(tpl.TitleTemplate)
	tpl.BodyTemplate = strings.TrimSpace(tpl.BodyTemplate)
	if tpl.TenantID == "" || tpl.Key == "" || !tpl.Channel.Valid() || tpl.TitleTemplate == "" || tpl.BodyTemplate == "" {
		return NotificationTemplate{}, fmt.Errorf("%w: tenant_id, key, channel, title_template and body_template are required", ErrInvalidRequest)
	}
	now := s.clock.Now().UTC()
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = now
	}
	tpl.UpdatedAt = now
	if err := s.repo.SaveTemplate(ctx, &tpl); err != nil {
		return NotificationTemplate{}, err
	}
	return tpl, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, tenantID string, id string) error {
	if s.repo == nil {
		return fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	id = strings.TrimSpace(id)
	if tenantID == "" || id == "" {
		return fmt.Errorf("%w: tenant_id and id are required", ErrInvalidRequest)
	}
	return s.repo.DeleteTemplate(ctx, tenantID, id)
}

func (s *Service) GetChannelConfig(ctx context.Context, tenantID string, channel Channel) (ChannelConfig, error) {
	if s.repo == nil {
		return ChannelConfig{}, fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ChannelConfig{}, fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}
	return s.repo.GetChannelConfig(ctx, tenantID, channel)
}

func (s *Service) SaveChannelConfig(ctx context.Context, cfg ChannelConfig) (ChannelConfig, error) {
	if s.repo == nil {
		return ChannelConfig{}, fmt.Errorf("notification service: repository is required")
	}
	cfg.TenantID = strings.TrimSpace(cfg.TenantID)
	if cfg.TenantID == "" || cfg.Channel == "" {
		return ChannelConfig{}, fmt.Errorf("%w: tenant_id and channel are required", ErrInvalidRequest)
	}
	now := s.clock.Now().UTC()
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now

	if err := s.repo.SaveChannelConfig(ctx, &cfg); err != nil {
		return ChannelConfig{}, err
	}
	return cfg, nil
}

func (s *Service) ListChannelConfigs(ctx context.Context, tenantID string) ([]ChannelConfig, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("notification service: repository is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}
	return s.repo.ListChannelConfigs(ctx, tenantID)
}

func normalizeCreateRequest(request CreateRequest) CreateRequest {
	request.TenantID = strings.TrimSpace(request.TenantID)
	request.Recipient = strings.TrimSpace(request.Recipient)
	request.Channel = Channel(strings.TrimSpace(string(request.Channel)))
	request.TemplateKey = strings.TrimSpace(request.TemplateKey)
	request.Title = strings.TrimSpace(request.Title)
	request.Body = strings.TrimSpace(request.Body)
	request.GroupKey = strings.TrimSpace(request.GroupKey)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	if request.Metadata == nil {
		request.Metadata = Metadata{}
	}
	return request
}

func validateCreateRequest(request CreateRequest) error {
	switch {
	case request.TenantID == "":
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	case request.Recipient == "":
		return fmt.Errorf("%w: recipient is required", ErrInvalidRequest)
	case !request.Channel.Valid():
		return fmt.Errorf("%w: channel is unsupported", ErrInvalidRequest)
	case request.Body == "":
		return fmt.Errorf("%w: body is required", ErrInvalidRequest)
	case len(request.Recipient) > 512:
		return fmt.Errorf("%w: recipient is too long", ErrInvalidRequest)
	case len(request.Title) > 256:
		return fmt.Errorf("%w: title is too long", ErrInvalidRequest)
	case len(request.Body) > 10000:
		return fmt.Errorf("%w: body is too long", ErrInvalidRequest)
	case len(request.Metadata) > 50:
		return fmt.Errorf("%w: metadata has too many entries", ErrInvalidRequest)
	default:
		return nil
	}
}

func mergeIntoAggregate(existing Notification, request CreateRequest, now time.Time) Notification {
	if existing.AggregateCount < 1 {
		existing.AggregateCount = 1
	}
	existing.TemplateKey = request.TemplateKey
	existing.Title = request.Title
	existing.Body = request.Body
	existing.IdempotencyKey = request.IdempotencyKey
	existing.Status = StatusQueued
	existing.AggregateCount++
	existing.UpdatedAt = now
	existing.Metadata = mergeMetadata(existing.Metadata, request.Metadata)
	return existing
}

func mergeMetadata(base Metadata, overlay Metadata) Metadata {
	merged := base.Clone()
	for key, value := range overlay {
		merged[key] = value
	}
	return merged
}

func (s *Service) renderTemplate(ctx context.Context, request *CreateRequest) error {
	if request.TemplateKey == "" || (request.Title != "" && request.Body != "") {
		return nil
	}
	if s.templates == nil {
		return nil
	}
	tpl, err := s.templates.GetTemplate(ctx, request.TenantID, request.TemplateKey, request.Channel)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: template not found", ErrInvalidRequest)
		}
		return err
	}
	data := map[string]any{
		"TenantID":  request.TenantID,
		"Recipient": request.Recipient,
		"Channel":   request.Channel,
		"Metadata":  request.Metadata,
	}
	if request.Title == "" {
		rendered, err := renderTextTemplate(tpl.TitleTemplate, data)
		if err != nil {
			return fmt.Errorf("%w: render title template", ErrInvalidRequest)
		}
		request.Title = rendered
	}
	if request.Body == "" {
		rendered, err := renderTextTemplate(tpl.BodyTemplate, data)
		if err != nil {
			return fmt.Errorf("%w: render body template", ErrInvalidRequest)
		}
		request.Body = rendered
	}
	return nil
}

func renderTextTemplate(source string, data any) (string, error) {
	tpl, err := template.New("notification").Option("missingkey=error").Parse(source)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}
