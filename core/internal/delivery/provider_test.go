package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

func TestDispatcherRoutesByChannel(t *testing.T) {
	repo := &fakeDeliveryRepository{
		configs: map[string]notification.ChannelConfig{
			"tenant-a:in_app": {
				Enabled: true,
				Config:  "{}",
			},
		},
	}
	dispatcher := NewConfiguredDispatcher(repo, &fakeInAppStore{id: "inapp_123"})

	got, err := dispatcher.Deliver(context.Background(), notification.Notification{
		ID:        "notif_1",
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   notification.ChannelInApp,
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "inapp:inapp_123" {
		t.Fatalf("provider message ID = %q, want inapp:inapp_123", got.ProviderMessageID)
	}
}

func TestDispatcherRejectsUnconfiguredChannel(t *testing.T) {
	var dispatcher *Dispatcher
	_, err := dispatcher.Deliver(context.Background(), notification.Notification{
		ID:      "notif_1",
		Channel: notification.ChannelEmail,
	})

	if !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("error = %v, want ErrUnsupportedChannel", err)
	}
}

func TestSMTPProviderBuildsAndSendsEmail(t *testing.T) {
	transport := &fakeSMTPTransport{}
	provider := NewSMTPProvider(EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     587,
		Username: "mailer",
		Password: "secret",
		From:     "Aetheris <noreply@example.com>",
		TLSMode:  "starttls",
		Timeout:  time.Second,
		Headers: map[string]string{
			"X-Product": "Aetheris",
		},
	}, transport)

	got, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_email",
		Recipient: "alice@example.com, bob@example.com",
		Channel:   notification.ChannelEmail,
		Title:     "Invoice ready",
		Body:      "Your invoice is ready.",
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "smtp:notif_email" {
		t.Fatalf("provider message ID = %q, want smtp:notif_email", got.ProviderMessageID)
	}
	if transport.config.Host != "smtp.example.com" || transport.config.TLSMode != "starttls" {
		t.Fatalf("smtp config = %#v", transport.config)
	}
	if len(transport.recipients) != 2 || transport.recipients[0] != "alice@example.com" || transport.recipients[1] != "bob@example.com" {
		t.Fatalf("recipients = %#v", transport.recipients)
	}
	message := string(transport.message)
	for _, want := range []string{
		"From: Aetheris <noreply@example.com>",
		"To: alice@example.com, bob@example.com",
		"Subject: Invoice ready",
		"X-Product: Aetheris",
		"X-Aetheris-Notification-ID: notif_email",
		"Your invoice is ready.",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("message missing %q:\n%s", want, message)
		}
	}
}

func TestHTTPProviderPostsTemplatedRequestAndExtractsResponseID(t *testing.T) {
	var body bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer sms-token" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		_, _ = body.ReadFrom(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message_id":"sms_123"}`))
	}))
	defer server.Close()

	provider := NewHTTPProvider(HTTPProviderConfig{
		Name:                "sms",
		Enabled:             true,
		URLTemplate:         server.URL + "/send",
		Method:              http.MethodPost,
		Headers:             map[string]string{"Authorization": "Bearer sms-token"},
		BodyTemplate:        `{"to":"{{ .Recipient }}","text":{{ quote .Body }},"tenant":"{{ .TenantID }}"}`,
		Timeout:             time.Second,
		SuccessStatusMin:    200,
		SuccessStatusMax:    299,
		ResponseIDJSONField: "message_id",
	}, server.Client())

	got, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_sms",
		TenantID:  "tenant-a",
		Recipient: "+15551234567",
		Channel:   notification.ChannelSMS,
		Body:      "Your code is 1234",
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "sms_123" {
		t.Fatalf("provider message ID = %q, want sms_123", got.ProviderMessageID)
	}
	var payload map[string]string
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		t.Fatalf("request body is not JSON: %v", err)
	}
	if payload["to"] != "+15551234567" || payload["text"] != "Your code is 1234" || payload["tenant"] != "tenant-a" {
		t.Fatalf("request payload = %#v", payload)
	}
}

func TestWebhookProviderRejectsPrivateIPByDefault(t *testing.T) {
	provider := NewWebhookProvider(WebhookConfig{
		Enabled:     true,
		URLTemplate: "{{ .Recipient }}",
		Method:      http.MethodPost,
	}, http.DefaultClient)

	_, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_webhook",
		Recipient: "http://127.0.0.1:8080/hook",
		Channel:   notification.ChannelWebhook,
	})

	if !errors.Is(err, ErrWebhookTargetNotAllowed) {
		t.Fatalf("error = %v, want ErrWebhookTargetNotAllowed", err)
	}
}

func TestWebhookProviderPostsSignedPayloadToAllowedHost(t *testing.T) {
	var signature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature = r.Header.Get("X-Aetheris-Signature")
		w.Header().Set("X-Delivery-ID", "webhook_123")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	provider := NewWebhookProvider(WebhookConfig{
		Enabled:          true,
		URLTemplate:      "{{ .Recipient }}",
		Method:           http.MethodPost,
		Headers:          map[string]string{"X-Static": "ok"},
		AllowedHosts:     []string{"127.0.0.1"},
		AllowPrivateIPs:  true,
		SigningSecret:    "hook-secret",
		ResponseIDHeader: "X-Delivery-ID",
		SuccessStatusMin: 200,
		SuccessStatusMax: 299,
	}, server.Client())

	got, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_webhook",
		Recipient: server.URL,
		Channel:   notification.ChannelWebhook,
		Title:     "Deploy finished",
		Body:      "Build 123 is green.",
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "webhook_123" {
		t.Fatalf("provider message ID = %q, want webhook_123", got.ProviderMessageID)
	}
	if !strings.HasPrefix(signature, "sha256=") {
		t.Fatalf("signature = %q, want sha256 prefix", signature)
	}
}

func TestInAppProviderStoresInboxMessage(t *testing.T) {
	store := &fakeInAppStore{id: "inapp_123"}
	provider := NewInAppProvider(store)

	got, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_inapp",
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   notification.ChannelInApp,
		Title:     "Welcome",
		Body:      "Hello",
		Metadata: notification.Metadata{
			"source": "onboarding",
		},
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "inapp:inapp_123" {
		t.Fatalf("provider message ID = %q, want inapp:inapp_123", got.ProviderMessageID)
	}
	if store.message.NotificationID != "notif_inapp" || store.message.UserID != "user-1" {
		t.Fatalf("stored message = %#v", store.message)
	}
	if store.message.Metadata["source"] != "onboarding" {
		t.Fatalf("stored metadata = %#v", store.message.Metadata)
	}
}

type fakeProvider struct {
	delivered notification.Notification
	result    jobs.DeliveryResult
	err       error
}

func (p *fakeProvider) Deliver(_ context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	p.delivered = record
	return p.result, p.err
}

type fakeSMTPTransport struct {
	config     EmailConfig
	from       string
	recipients []string
	message    []byte
	err        error
}

func (t *fakeSMTPTransport) Send(_ context.Context, config EmailConfig, from string, recipients []string, message []byte) error {
	t.config = config
	t.from = from
	t.recipients = recipients
	t.message = message
	return t.err
}

type fakeInAppStore struct {
	id      string
	message notification.InAppMessage
	err     error
}

func (s *fakeInAppStore) CreateInAppMessage(_ context.Context, message *notification.InAppMessage) error {
	if s.err != nil {
		return s.err
	}
	message.ID = s.id
	s.message = *message
	return nil
}
