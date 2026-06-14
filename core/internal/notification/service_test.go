package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestServiceCreateNotificationPersistsAndEnqueuesDelivery(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepository{}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: now})

	got, err := service.Create(context.Background(), CreateRequest{
		TenantID:    "tenant-a",
		Recipient:   "user-1",
		Channel:     ChannelEmail,
		TemplateKey: "billing.invoice.ready",
		Title:       "Invoice ready",
		Body:        "Your invoice is ready.",
		GroupKey:    "billing:user-1",
		Metadata: Metadata{
			"invoice_id": "inv_123",
		},
	})

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID == "" {
		t.Fatal("Create should assign an ID")
	}
	if got.Status != StatusQueued {
		t.Fatalf("status = %q, want %q", got.Status, StatusQueued)
	}
	if got.AggregateCount != 1 {
		t.Fatalf("aggregate count = %d, want 1", got.AggregateCount)
	}
	if got.CreatedAt != now || got.UpdatedAt != now {
		t.Fatalf("timestamps = %s/%s, want %s", got.CreatedAt, got.UpdatedAt, now)
	}
	if !repo.created {
		t.Fatal("repository Create was not called")
	}
	if queue.notificationID != got.ID {
		t.Fatalf("queued notification ID = %q, want %q", queue.notificationID, got.ID)
	}
}

func TestServiceCreateNotificationAggregatesOpenNotification(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 30, 0, 0, time.UTC)
	existing := &Notification{
		ID:             "notif_existing",
		TenantID:       "tenant-a",
		Recipient:      "user-1",
		Channel:        ChannelInApp,
		TemplateKey:    "system.digest",
		Title:          "Earlier title",
		Body:           "Earlier body",
		GroupKey:       "digest:user-1",
		Status:         StatusQueued,
		AggregateCount: 2,
		Metadata: Metadata{
			"first": "true",
		},
		CreatedAt: now.Add(-time.Minute),
		UpdatedAt: now.Add(-time.Minute),
	}
	repo := &fakeRepository{existing: existing}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: now})

	got, err := service.Create(context.Background(), CreateRequest{
		TenantID:    "tenant-a",
		Recipient:   "user-1",
		Channel:     ChannelInApp,
		TemplateKey: "system.digest",
		Title:       "Latest title",
		Body:        "Latest body",
		GroupKey:    "digest:user-1",
		Metadata: Metadata{
			"latest": "true",
		},
	})

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("aggregated ID = %q, want %q", got.ID, existing.ID)
	}
	if got.AggregateCount != 3 {
		t.Fatalf("aggregate count = %d, want 3", got.AggregateCount)
	}
	if got.Title != "Latest title" || got.Body != "Latest body" {
		t.Fatalf("title/body were not refreshed: %q/%q", got.Title, got.Body)
	}
	if got.Metadata["first"] != "true" || got.Metadata["latest"] != "true" {
		t.Fatalf("metadata was not merged: %#v", got.Metadata)
	}
	if !repo.updated || repo.created {
		t.Fatalf("expected update-only aggregate path, created=%v updated=%v", repo.created, repo.updated)
	}
	if queue.notificationID != existing.ID {
		t.Fatalf("queued notification ID = %q, want %q", queue.notificationID, existing.ID)
	}
}

func TestServiceCreateNotificationRejectsInvalidInput(t *testing.T) {
	repo := &fakeRepository{}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: time.Now()})

	_, err := service.Create(context.Background(), CreateRequest{
		TenantID:  "",
		Recipient: "user-1",
		Channel:   ChannelEmail,
		Title:     "Missing tenant",
		Body:      "This should fail.",
	})

	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
	if repo.created || repo.updated || queue.notificationID != "" {
		t.Fatalf("invalid input should not persist or enqueue, repo=%#v queue=%#v", repo, queue)
	}
}

func TestServiceCreateNotificationRejectsUnsupportedChannel(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeQueue{}, fixedClock{now: time.Now()})

	_, err := service.Create(context.Background(), CreateRequest{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   Channel("pager"),
		Title:     "Bad channel",
		Body:      "This should fail.",
	})

	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestServiceCreateNotificationPropagatesQueueError(t *testing.T) {
	queueErr := errors.New("redis unavailable")
	service := NewService(&fakeRepository{}, &fakeQueue{err: queueErr}, fixedClock{now: time.Now()})

	_, err := service.Create(context.Background(), CreateRequest{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelEmail,
		Title:     "Queued",
		Body:      "This should fail after persistence.",
	})

	if !errors.Is(err, queueErr) {
		t.Fatalf("error = %v, want queue error", err)
	}
}

func TestServiceCreateNotificationReturnsExistingForIdempotencyKey(t *testing.T) {
	existing := &Notification{
		ID:             "notif_existing",
		TenantID:       "tenant-a",
		Recipient:      "user-1",
		Channel:        ChannelInApp,
		Title:          "Existing",
		Body:           "Existing body",
		Status:         StatusQueued,
		IdempotencyKey: "request-123",
	}
	repo := &fakeRepository{idempotent: existing}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: time.Now()})

	got, err := service.Create(context.Background(), CreateRequest{
		TenantID:       "tenant-a",
		Recipient:      "user-1",
		Channel:        ChannelInApp,
		Title:          "New",
		Body:           "New body",
		IdempotencyKey: "request-123",
	})

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("ID = %q, want existing %q", got.ID, existing.ID)
	}
	if repo.created || queue.notificationID != "" {
		t.Fatalf("idempotent request should not create or enqueue, created=%v queued=%q", repo.created, queue.notificationID)
	}
}

func TestServiceCreateNotificationRendersTemplate(t *testing.T) {
	repo := &fakeRepository{}
	queue := &fakeQueue{}
	service := NewServiceWithOptions(repo, queue, fixedClock{now: time.Now()}, ServiceOptions{
		Templates: fakeTemplateStore{
			template: NotificationTemplate{
				TenantID:      "tenant-a",
				Key:           "welcome",
				Channel:       ChannelInApp,
				TitleTemplate: "Welcome {{ .Metadata.name }}",
				BodyTemplate:  "Hello {{ .Recipient }} from {{ .TenantID }}",
			},
		},
	})

	got, err := service.Create(context.Background(), CreateRequest{
		TenantID:    "tenant-a",
		Recipient:   "user-1",
		Channel:     ChannelInApp,
		TemplateKey: "welcome",
		Metadata: Metadata{
			"name": "Alice",
		},
	})

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.Title != "Welcome Alice" || got.Body != "Hello user-1 from tenant-a" {
		t.Fatalf("rendered title/body = %q/%q", got.Title, got.Body)
	}
}

func TestServiceListAndInAppMethodsDelegateToRepository(t *testing.T) {
	repo := &fakeRepository{
		list:  []Notification{{ID: "notif_1", TenantID: "tenant-a"}},
		inbox: []InAppMessage{{ID: "msg_1", TenantID: "tenant-a", UserID: "user-1"}},
	}
	service := NewService(repo, &fakeQueue{}, fixedClock{now: time.Now()})

	list, err := service.List(context.Background(), NotificationQuery{TenantID: "tenant-a", Limit: 5})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 1 || repo.query.TenantID != "tenant-a" || repo.query.Limit != 5 {
		t.Fatalf("list=%#v query=%#v", list, repo.query)
	}

	inbox, err := service.ListInApp(context.Background(), InAppQuery{TenantID: "tenant-a", UserID: "user-1", UnreadOnly: true})
	if err != nil {
		t.Fatalf("ListInApp returned error: %v", err)
	}
	if len(inbox) != 1 || !repo.inboxQuery.UnreadOnly {
		t.Fatalf("inbox=%#v query=%#v", inbox, repo.inboxQuery)
	}

	if err := service.MarkInAppRead(context.Background(), "tenant-a", "msg_1", "user-1"); err != nil {
		t.Fatalf("MarkInAppRead returned error: %v", err)
	}
	if repo.markTenantID != "tenant-a" || repo.markMessageID != "msg_1" || repo.markUserID != "user-1" {
		t.Fatalf("mark args = tenant:%q message:%q user:%q", repo.markTenantID, repo.markMessageID, repo.markUserID)
	}
}

func TestServiceManagementMethodsDelegateToRepository(t *testing.T) {
	repo := &fakeRepository{
		detail:    Notification{ID: "notif_1", TenantID: "tenant-a"},
		attempts:  []DeliveryAttempt{{ID: "attempt_1", TenantID: "tenant-a"}},
		templates: []NotificationTemplate{{ID: "tpl_1", TenantID: "tenant-a", Key: "welcome"}},
	}
	service := NewService(repo, &fakeQueue{}, fixedClock{now: time.Now()})

	detail, err := service.Get(context.Background(), "tenant-a", "notif_1")
	if err != nil || detail.ID != "notif_1" {
		t.Fatalf("Get = %#v, %v", detail, err)
	}

	attempts, err := service.ListDeliveryAttempts(context.Background(), "tenant-a", "notif_1")
	if err != nil || len(attempts) != 1 {
		t.Fatalf("attempts = %#v, %v", attempts, err)
	}

	templates, err := service.ListTemplates(context.Background(), TemplateQuery{TenantID: "tenant-a"})
	if err != nil || len(templates) != 1 {
		t.Fatalf("templates = %#v, %v", templates, err)
	}

	tpl, err := service.SaveTemplate(context.Background(), NotificationTemplate{
		TenantID:      "tenant-a",
		Key:           "welcome",
		Channel:       ChannelInApp,
		TitleTemplate: "Welcome",
		BodyTemplate:  "Hello",
	})
	if err != nil || tpl.ID == "" {
		t.Fatalf("SaveTemplate = %#v, %v", tpl, err)
	}

	if err := service.DeleteTemplate(context.Background(), "tenant-a", "tpl_1"); err != nil {
		t.Fatalf("DeleteTemplate returned error: %v", err)
	}
	if repo.deletedTemplateTenantID != "tenant-a" || repo.deletedTemplateID != "tpl_1" {
		t.Fatalf("delete args = tenant:%q id:%q", repo.deletedTemplateTenantID, repo.deletedTemplateID)
	}
}

func TestServiceCreateNotificationRequiresDependencies(t *testing.T) {
	request := CreateRequest{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelEmail,
		Title:     "Dependency check",
		Body:      "This should fail before persistence.",
	}

	_, err := NewService(nil, &fakeQueue{}, fixedClock{now: time.Now()}).Create(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), "repository") {
		t.Fatalf("error = %v, want repository dependency error", err)
	}

	_, err = NewService(&fakeRepository{}, nil, fixedClock{now: time.Now()}).Create(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), "delivery queue") {
		t.Fatalf("error = %v, want queue dependency error", err)
	}
}

func TestSystemClockReturnsCurrentUTC(t *testing.T) {
	got := SystemClock{}.Now()

	if got.Location() != time.UTC {
		t.Fatalf("location = %s, want UTC", got.Location())
	}
	if time.Since(got) > time.Second {
		t.Fatalf("clock returned stale time %s", got)
	}
}

func TestValidateCreateRequestCoversRequiredFields(t *testing.T) {
	base := CreateRequest{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelEmail,
		Title:     "Hello",
		Body:      "World",
	}

	cases := []CreateRequest{
		{Recipient: base.Recipient, Channel: base.Channel, Title: base.Title, Body: base.Body},
		{TenantID: base.TenantID, Channel: base.Channel, Title: base.Title, Body: base.Body},
		{TenantID: base.TenantID, Recipient: base.Recipient, Channel: Channel("bad"), Title: base.Title, Body: base.Body},
		{TenantID: base.TenantID, Recipient: base.Recipient, Channel: base.Channel, Title: base.Title},
	}

	for _, request := range cases {
		if err := validateCreateRequest(request); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("validateCreateRequest(%#v) = %v, want ErrInvalidRequest", request, err)
		}
	}
}

func TestServiceCreateNotificationFailsWhenRecipientIsMissingAndNotConfigured(t *testing.T) {
	repo := &fakeRepository{}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: time.Now()})

	_, err := service.Create(context.Background(), CreateRequest{
		TenantID: "tenant-a",
		Channel:  ChannelSMS,
		Body:     "Test message",
	})

	if !errors.Is(err, ErrInvalidRequest) || !strings.Contains(err.Error(), "recipient is required") {
		t.Fatalf("error = %v, want ErrInvalidRequest with 'recipient is required'", err)
	}
}

func TestServiceCreateNotificationSucceedsWhenRecipientIsMissingButConfigured(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepository{
		configs: map[string]ChannelConfig{
			"tenant-a:sms": {
				TenantID: "tenant-a",
				Channel:  ChannelSMS,
				Enabled:  true,
				Config:   `{"default_recipient": "config-recipient"}`,
			},
		},
	}
	queue := &fakeQueue{}
	service := NewService(repo, queue, fixedClock{now: now})

	got, err := service.Create(context.Background(), CreateRequest{
		TenantID: "tenant-a",
		Channel:  ChannelSMS,
		Body:     "Test message",
	})

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.Recipient != "config-recipient" {
		t.Fatalf("recipient = %q, want 'config-recipient'", got.Recipient)
	}
	if got.Title != "" {
		t.Fatalf("title = %q, want empty string", got.Title)
	}
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type fakeRepository struct {
	existing                *Notification
	idempotent              *Notification
	created                 bool
	updated                 bool
	list                    []Notification
	query                   NotificationQuery
	inbox                   []InAppMessage
	inboxQuery              InAppQuery
	detail                  Notification
	attempts                []DeliveryAttempt
	templates               []NotificationTemplate
	templateQuery           TemplateQuery
	savedTemplate           NotificationTemplate
	deletedTemplateTenantID string
	deletedTemplateID       string
	markTenantID            string
	markMessageID           string
	markUserID              string
	configs                 map[string]ChannelConfig
}

func (r *fakeRepository) FindByIdempotencyKey(_ context.Context, _ string, _ string) (*Notification, error) {
	return r.idempotent, nil
}

func (r *fakeRepository) FindOpenAggregate(_ context.Context, _ AggregateKey) (*Notification, error) {
	return r.existing, nil
}

func (r *fakeRepository) Create(_ context.Context, notification *Notification) error {
	r.created = true
	return nil
}

func (r *fakeRepository) UpdateAggregate(_ context.Context, notification *Notification) error {
	r.updated = true
	if r.existing != nil {
		*r.existing = *notification
	}
	return nil
}

func (r *fakeRepository) List(_ context.Context, query NotificationQuery) ([]Notification, error) {
	r.query = query
	return r.list, nil
}

func (r *fakeRepository) GetByTenantID(_ context.Context, tenantID string, id string) (Notification, error) {
	if r.detail.TenantID == tenantID && r.detail.ID == id {
		return r.detail, nil
	}
	return Notification{}, ErrNotFound
}

func (r *fakeRepository) ListDeliveryAttempts(_ context.Context, tenantID string, notificationID string) ([]DeliveryAttempt, error) {
	return r.attempts, nil
}

func (r *fakeRepository) ListInApp(_ context.Context, query InAppQuery) ([]InAppMessage, error) {
	r.inboxQuery = query
	return r.inbox, nil
}

func (r *fakeRepository) MarkInAppRead(_ context.Context, tenantID string, messageID string, userID string, _ time.Time) error {
	r.markTenantID = tenantID
	r.markMessageID = messageID
	r.markUserID = userID
	return nil
}

func (r *fakeRepository) ListTemplates(_ context.Context, query TemplateQuery) ([]NotificationTemplate, error) {
	r.templateQuery = query
	return r.templates, nil
}

func (r *fakeRepository) SaveTemplate(_ context.Context, template *NotificationTemplate) error {
	if template.ID == "" {
		template.ID = "tpl_1"
	}
	r.savedTemplate = *template
	return nil
}

func (r *fakeRepository) DeleteTemplate(_ context.Context, tenantID string, id string) error {
	r.deletedTemplateTenantID = tenantID
	r.deletedTemplateID = id
	return nil
}

func (r *fakeRepository) GetChannelConfig(_ context.Context, tenantID string, channel Channel) (ChannelConfig, error) {
	if r.configs == nil {
		return ChannelConfig{}, ErrNotFound
	}
	key := tenantID + ":" + string(channel)
	if cfg, ok := r.configs[key]; ok {
		return cfg, nil
	}
	return ChannelConfig{}, ErrNotFound
}

func (r *fakeRepository) SaveChannelConfig(_ context.Context, config *ChannelConfig) error {
	if r.configs == nil {
		r.configs = make(map[string]ChannelConfig)
	}
	if config.ID == "" {
		config.ID = "cfg_1"
	}
	key := config.TenantID + ":" + string(config.Channel)
	r.configs[key] = *config
	return nil
}

func (r *fakeRepository) ListChannelConfigs(_ context.Context, tenantID string) ([]ChannelConfig, error) {
	var list []ChannelConfig
	for _, cfg := range r.configs {
		if cfg.TenantID == tenantID {
			list = append(list, cfg)
		}
	}
	return list, nil
}

func TestServiceChannelConfigMethods(t *testing.T) {
	repo := &fakeRepository{
		configs: map[string]ChannelConfig{
			"tenant-a:email": {
				ID:       "cfg_1",
				TenantID: "tenant-a",
				Channel:  ChannelEmail,
				Enabled:  true,
				Config:   `{"host":"smtp.mail.com"}`,
			},
		},
	}
	service := NewService(repo, &fakeQueue{}, fixedClock{now: time.Now()})

	// Get
	cfg, err := service.GetChannelConfig(context.Background(), "tenant-a", ChannelEmail)
	if err != nil || cfg.ID != "cfg_1" {
		t.Fatalf("GetChannelConfig = %+v, err = %v", cfg, err)
	}

	// List
	list, err := service.ListChannelConfigs(context.Background(), "tenant-a")
	if err != nil || len(list) != 1 {
		t.Fatalf("ListChannelConfigs = %+v, err = %v", list, err)
	}

	// Save
	newCfg := ChannelConfig{
		TenantID: "tenant-a",
		Channel:  ChannelSMS,
		Enabled:  true,
		Config:   `{"url":"http://sms"}`,
	}
	saved, err := service.SaveChannelConfig(context.Background(), newCfg)
	if err != nil || saved.ID == "" || saved.Channel != ChannelSMS {
		t.Fatalf("SaveChannelConfig = %+v, err = %v", saved, err)
	}
}

type fakeTemplateStore struct {
	template NotificationTemplate
}

func (s fakeTemplateStore) GetTemplate(_ context.Context, tenantID string, key string, channel Channel) (NotificationTemplate, error) {
	if s.template.TenantID == tenantID && s.template.Key == key && s.template.Channel == channel {
		return s.template, nil
	}
	return NotificationTemplate{}, ErrNotFound
}

type fakeQueue struct {
	notificationID string
	err            error
}

func (q *fakeQueue) EnqueueDelivery(_ context.Context, notification Notification) error {
	if q.err != nil {
		return q.err
	}
	q.notificationID = notification.ID
	return nil
}
