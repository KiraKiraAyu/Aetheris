package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"aetheris/internal/notification"
)

func TestCreateNotificationRouteReturnsAcceptedNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{
		notification: notification.Notification{
			ID:             "notif_1",
			TenantID:       "tenant-a",
			Recipient:      "user-1",
			Channel:        notification.ChannelEmail,
			TemplateKey:    "billing.invoice.ready",
			Title:          "Invoice ready",
			Body:           "Your invoice is ready.",
			GroupKey:       "billing:user-1",
			Status:         notification.StatusQueued,
			AggregateCount: 1,
			CreatedAt:      time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC),
		},
	}
	router := gin.New()
	RegisterRoutes(router, service)

	body := bytes.NewBufferString(`{
		"tenant_id": "tenant-a",
		"recipient": "user-1",
		"channel": "email",
		"template_key": "billing.invoice.ready",
		"title": "Invoice ready",
		"body": "Your invoice is ready.",
		"group_key": "billing:user-1",
		"metadata": {"invoice_id": "inv_123"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/send", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if service.request.TenantID != "tenant-a" || service.request.Channel != notification.ChannelEmail {
		t.Fatalf("service request was not decoded correctly: %#v", service.request)
	}
	var response notification.Notification
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not a notification: %v", err)
	}
	if response.ID != "notif_1" || response.Status != notification.StatusQueued {
		t.Fatalf("response = %#v, want queued notif_1", response)
	}
}

func TestCreateNotificationRouteMapsValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{err: notification.ErrInvalidRequest}
	router := gin.New()
	RegisterRoutes(router, service)

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{"tenant_id": ""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !service.called {
		t.Fatal("service should receive syntactically valid JSON and return domain validation")
	}
}

func TestCreateNotificationRouteRejectsMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{}
	router := gin.New()
	RegisterRoutes(router, service)

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if service.called {
		t.Fatal("service should not be called for malformed JSON")
	}
}

func TestAuthenticatedCreateNotificationUsesTenantFromAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{
		notification: notification.Notification{ID: "notif_1", TenantID: "tenant-a", Status: notification.StatusQueued},
	}
	router := gin.New()
	RegisterRoutesWithOptions(router, service, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{
		"recipient": "user-1",
		"channel": "in_app",
		"title": "Hello",
		"body": "World"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-a")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if service.request.TenantID != "tenant-a" {
		t.Fatalf("tenant = %q, want tenant-a", service.request.TenantID)
	}
}

func TestAuthenticatedCreateNotificationRejectsMissingOrMismatchedTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &fakeNotificationService{}
	RegisterRoutesWithOptions(router, service, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing auth status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{
		"tenant_id": "tenant-b",
		"recipient": "user-1",
		"channel": "in_app",
		"title": "Hello",
		"body": "World"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("tenant mismatch status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if service.called {
		t.Fatal("service should not be called for tenant mismatch")
	}
}

func TestQueryRoutesAreTenantScoped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{
		list: []notification.Notification{{ID: "notif_1", TenantID: "tenant-a"}},
		inbox: []notification.InAppMessage{{
			ID:       "msg_1",
			TenantID: "tenant-a",
			UserID:   "user-1",
		}},
	}
	router := gin.New()
	RegisterRoutesWithOptions(router, service, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	req := httptest.NewRequest(http.MethodGet, "/notifications?recipient=user-1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer secret-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list notifications status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.query.TenantID != "tenant-a" || service.query.Recipient != "user-1" || service.query.Limit != 10 {
		t.Fatalf("notification query = %#v", service.query)
	}

	req = httptest.NewRequest(http.MethodGet, "/in-app/messages?user_id=user-1&unread=true", nil)
	req.Header.Set("Authorization", "Bearer secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list inbox status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.inboxQuery.TenantID != "tenant-a" || !service.inboxQuery.UnreadOnly {
		t.Fatalf("inbox query = %#v", service.inboxQuery)
	}

	req = httptest.NewRequest(http.MethodPost, "/in-app/messages/msg_1/read?user_id=user-1", nil)
	req.Header.Set("Authorization", "Bearer secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("mark read status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if service.markTenantID != "tenant-a" || service.markMessageID != "msg_1" || service.markUserID != "user-1" {
		t.Fatalf("mark read args = tenant:%q message:%q user:%q", service.markTenantID, service.markMessageID, service.markUserID)
	}
}

func TestManagementRoutesExposeDetailsAttemptsAndTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{
		notification: notification.Notification{ID: "notif_1", TenantID: "tenant-a"},
		attempts: []notification.DeliveryAttempt{{
			ID:             "attempt_1",
			NotificationID: "notif_1",
			TenantID:       "tenant-a",
			Status:         notification.AttemptStatusDelivered,
		}},
		templates: []notification.NotificationTemplate{{
			ID:       "tpl_1",
			TenantID: "tenant-a",
			Key:      "welcome",
			Channel:  notification.ChannelInApp,
		}},
	}
	router := gin.New()
	RegisterRoutesWithOptions(router, service, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	req := httptest.NewRequest(http.MethodGet, "/notifications/notif_1", nil)
	req.Header.Set("X-API-Key", "secret-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.detailTenantID != "tenant-a" || service.detailID != "notif_1" {
		t.Fatalf("detail args = tenant:%q id:%q", service.detailTenantID, service.detailID)
	}

	req = httptest.NewRequest(http.MethodGet, "/notifications/notif_1/attempts", nil)
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("attempts status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.attemptTenantID != "tenant-a" || service.attemptNotificationID != "notif_1" {
		t.Fatalf("attempt args = tenant:%q notification:%q", service.attemptTenantID, service.attemptNotificationID)
	}

	req = httptest.NewRequest(http.MethodGet, "/templates?channel=in_app", nil)
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("templates status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.templateQuery.TenantID != "tenant-a" || service.templateQuery.Channel != notification.ChannelInApp {
		t.Fatalf("template query = %#v", service.templateQuery)
	}

	req = httptest.NewRequest(http.MethodPost, "/templates", bytes.NewBufferString(`{
		"key": "digest",
		"channel": "email",
		"title_template": "Digest",
		"body_template": "Hello {{ .Recipient }}"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create template status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if service.template.TenantID != "tenant-a" || service.template.Key != "digest" {
		t.Fatalf("created template = %#v", service.template)
	}

	req = httptest.NewRequest(http.MethodDelete, "/templates/tpl_1", nil)
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete template status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if service.deleteTemplateTenantID != "tenant-a" || service.deleteTemplateID != "tpl_1" {
		t.Fatalf("delete args = tenant:%q id:%q", service.deleteTemplateTenantID, service.deleteTemplateID)
	}
}

func TestRateLimitMiddlewareReturnsTooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutesWithOptions(router, &fakeNotificationService{}, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
		RateLimiter:   fakeRateLimiter{allowed: false},
	})

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Authorization", "Bearer secret-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestCORSMiddlewareHandlesPreflightBeforeAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutesWithOptions(router, &fakeNotificationService{}, Options{
		AllowedOrigins: []string{"http://127.0.0.1:5178"},
		Authenticator:  NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	req := httptest.NewRequest(http.MethodOptions, "/notifications", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5178")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:5178" {
		t.Fatalf("allow origin = %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("allow credentials = %q", rec.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestRoutesMapServiceErrorsAndBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutesWithOptions(router, &fakeNotificationService{err: notification.ErrNotFound}, Options{MaxBodyBytes: 8})

	req := httptest.NewRequest(http.MethodGet, "/notifications?tenant_id=tenant-a", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	req = httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(`{"tenant_id":"tenant-a"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("body limit status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestStaticAPIKeyAuthenticatorRejectsBadTokens(t *testing.T) {
	authenticator := NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic secret-a")

	if _, ok := authenticator.Authenticate(req); ok {
		t.Fatal("basic auth token should not authenticate")
	}
}

type fakeNotificationService struct {
	called                   bool
	request                  notification.CreateRequest
	notification             notification.Notification
	list                     []notification.Notification
	query                    notification.NotificationQuery
	detailTenantID           string
	detailID                 string
	attempts                 []notification.DeliveryAttempt
	attemptTenantID          string
	attemptNotificationID    string
	inbox                    []notification.InAppMessage
	inboxQuery               notification.InAppQuery
	template                 notification.NotificationTemplate
	templates                []notification.NotificationTemplate
	templateQuery            notification.TemplateQuery
	deleteTemplateTenantID   string
	deleteTemplateID         string
	markTenantID             string
	markMessageID            string
	markUserID               string
	channelConfigs           []notification.ChannelConfig
	listChannelConfigsTenant string
	savedChannelConfig       notification.ChannelConfig
	err                      error
}

func (s *fakeNotificationService) Create(_ context.Context, request notification.CreateRequest) (notification.Notification, error) {
	s.called = true
	s.request = request
	if s.err != nil {
		if errors.Is(s.err, notification.ErrInvalidRequest) {
			return notification.Notification{}, s.err
		}
		return notification.Notification{}, s.err
	}
	return s.notification, nil
}

func (s *fakeNotificationService) List(ctx context.Context, query notification.NotificationQuery) ([]notification.Notification, error) {
	s.query = query
	return s.list, s.err
}

func (s *fakeNotificationService) Get(ctx context.Context, tenantID string, id string) (notification.Notification, error) {
	s.detailTenantID = tenantID
	s.detailID = id
	return s.notification, s.err
}

func (s *fakeNotificationService) ListDeliveryAttempts(ctx context.Context, tenantID string, notificationID string) ([]notification.DeliveryAttempt, error) {
	s.attemptTenantID = tenantID
	s.attemptNotificationID = notificationID
	return s.attempts, s.err
}

func (s *fakeNotificationService) ListInApp(ctx context.Context, query notification.InAppQuery) ([]notification.InAppMessage, error) {
	s.inboxQuery = query
	return s.inbox, s.err
}

func (s *fakeNotificationService) MarkInAppRead(ctx context.Context, tenantID string, messageID string, userID string) error {
	s.markTenantID = tenantID
	s.markMessageID = messageID
	s.markUserID = userID
	return s.err
}

func (s *fakeNotificationService) ListTemplates(ctx context.Context, query notification.TemplateQuery) ([]notification.NotificationTemplate, error) {
	s.templateQuery = query
	return s.templates, s.err
}

func (s *fakeNotificationService) SaveTemplate(ctx context.Context, template notification.NotificationTemplate) (notification.NotificationTemplate, error) {
	s.template = template
	if s.template.ID == "" {
		s.template.ID = "tpl_new"
	}
	return s.template, s.err
}

func (s *fakeNotificationService) DeleteTemplate(ctx context.Context, tenantID string, id string) error {
	s.deleteTemplateTenantID = tenantID
	s.deleteTemplateID = id
	return s.err
}

func (s *fakeNotificationService) ListChannelConfigs(ctx context.Context, tenantID string) ([]notification.ChannelConfig, error) {
	s.listChannelConfigsTenant = tenantID
	return s.channelConfigs, s.err
}

func (s *fakeNotificationService) SaveChannelConfig(ctx context.Context, cfg notification.ChannelConfig) (notification.ChannelConfig, error) {
	s.savedChannelConfig = cfg
	if s.savedChannelConfig.ID == "" {
		s.savedChannelConfig.ID = "cfg_new"
	}
	return s.savedChannelConfig, s.err
}

func TestChannelConfigRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeNotificationService{
		channelConfigs: []notification.ChannelConfig{{
			ID:       "cfg_1",
			TenantID: "tenant-a",
			Channel:  notification.ChannelEmail,
			Enabled:  true,
			Config:   `{"host":"smtp.example.com"}`,
		}},
	}
	router := gin.New()
	RegisterRoutesWithOptions(router, service, Options{
		Authenticator: NewStaticAPIKeyAuthenticator(map[string]string{"secret-a": "tenant-a"}),
	})

	// GET /channels
	req := httptest.NewRequest(http.MethodGet, "/channels", nil)
	req.Header.Set("X-API-Key", "secret-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get channels status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.listChannelConfigsTenant != "tenant-a" {
		t.Fatalf("list tenant = %q, want tenant-a", service.listChannelConfigsTenant)
	}
	var configs []notification.ChannelConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &configs); err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if len(configs) != 1 || configs[0].ID != "cfg_1" {
		t.Fatalf("configs = %+v, want cfg_1", configs)
	}

	// POST /channels
	body := bytes.NewBufferString(`{
		"channel": "email",
		"enabled": true,
		"config": "{\"host\":\"smtp.gmail.com\"}"
	}`)
	req = httptest.NewRequest(http.MethodPost, "/channels", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret-a")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("post channel status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if service.savedChannelConfig.TenantID != "tenant-a" || service.savedChannelConfig.Channel != notification.ChannelEmail {
		t.Fatalf("savedChannelConfig = %+v, want tenant-a and email", service.savedChannelConfig)
	}
}

type fakeRateLimiter struct {
	allowed bool
}

func (l fakeRateLimiter) Allow(context.Context, string) (bool, error) {
	return l.allowed, nil
}
