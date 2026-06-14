package delivery

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type fakeDeliveryRepository struct {
	configs map[string]notification.ChannelConfig
}

func (r *fakeDeliveryRepository) GetByID(context.Context, string) (notification.Notification, error) {
	return notification.Notification{}, nil
}
func (r *fakeDeliveryRepository) MarkDeliveryResult(context.Context, string, notification.DeliveryUpdate) error {
	return nil
}
func (r *fakeDeliveryRepository) CountDeliveryAttempts(context.Context, string) (int, error) {
	return 0, nil
}
func (r *fakeDeliveryRepository) CreateDeliveryAttempt(context.Context, *notification.DeliveryAttempt) error {
	return nil
}
func (r *fakeDeliveryRepository) FinishDeliveryAttempt(context.Context, string, notification.DeliveryAttemptUpdate) error {
	return nil
}
func (r *fakeDeliveryRepository) GetChannelConfig(ctx context.Context, tenantID string, channel notification.Channel) (notification.ChannelConfig, error) {
	key := tenantID + ":" + string(channel)
	if cfg, ok := r.configs[key]; ok {
		return cfg, nil
	}
	return notification.ChannelConfig{}, notification.ErrNotFound
}

func TestNewConfiguredDispatcherBuildsEnabledProviders(t *testing.T) {
	repo := &fakeDeliveryRepository{
		configs: map[string]notification.ChannelConfig{
			"tenant-a:in_app": {
				Enabled: true,
				Config:  "{}",
			},
		},
	}
	dispatcher := NewConfiguredDispatcher(repo, &fakeInAppStore{id: "inapp_1"})

	got, err := dispatcher.Deliver(context.Background(), notification.Notification{
		ID:        "notif_1",
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   notification.ChannelInApp,
		Title:     "Hello",
		Body:      "Welcome",
	})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "inapp:inapp_1" {
		t.Fatalf("provider message ID = %q, want inapp:inapp_1", got.ProviderMessageID)
	}
}

func TestNewConfiguredDispatcherRejectsInvalidConfiguration(t *testing.T) {
	// 1. Channel not configured (ErrNotFound)
	repo := &fakeDeliveryRepository{configs: map[string]notification.ChannelConfig{}}
	dispatcher := NewConfiguredDispatcher(repo, nil)
	_, err := dispatcher.Deliver(context.Background(), notification.Notification{
		TenantID: "tenant-a",
		Channel:  notification.ChannelEmail,
	})
	if err == nil || !strings.Contains(err.Error(), "is not configured") {
		t.Fatalf("error = %v, want not configured error", err)
	}

	// 2. Channel disabled
	repo = &fakeDeliveryRepository{
		configs: map[string]notification.ChannelConfig{
			"tenant-a:email": {Enabled: false},
		},
	}
	dispatcher = NewConfiguredDispatcher(repo, nil)
	_, err = dispatcher.Deliver(context.Background(), notification.Notification{
		TenantID: "tenant-a",
		Channel:  notification.ChannelEmail,
	})
	if err == nil || !strings.Contains(err.Error(), "is disabled") {
		t.Fatalf("error = %v, want disabled error", err)
	}

	// 3. Email invalid config (missing host/from)
	repo = &fakeDeliveryRepository{
		configs: map[string]notification.ChannelConfig{
			"tenant-a:email": {Enabled: true, Config: `{"host":""}`},
		},
	}
	dispatcher = NewConfiguredDispatcher(repo, nil)
	_, err = dispatcher.Deliver(context.Background(), notification.Notification{
		TenantID: "tenant-a",
		Channel:  notification.ChannelEmail,
	})
	if err == nil || !strings.Contains(err.Error(), "host and from address are required") {
		t.Fatalf("error = %v, want email validation error", err)
	}

	// 4. InApp missing store
	repo = &fakeDeliveryRepository{
		configs: map[string]notification.ChannelConfig{
			"tenant-a:in_app": {Enabled: true, Config: `{}`},
		},
	}
	dispatcher = NewConfiguredDispatcher(repo, nil)
	_, err = dispatcher.Deliver(context.Background(), notification.Notification{
		TenantID: "tenant-a",
		Channel:  notification.ChannelInApp,
	})
	if err == nil || !strings.Contains(err.Error(), "store is required") {
		t.Fatalf("error = %v, want store is required error", err)
	}
}

func TestTelegramProviderRejectsInvalidInputAndPermanentFailures(t *testing.T) {
	_, err := NewTelegramProvider(TelegramConfig{Enabled: true}, nil).Deliver(context.Background(), notification.Notification{Recipient: "123"})
	if err == nil || !strings.Contains(err.Error(), "bot token") {
		t.Fatalf("error = %v, want bot token error", err)
	}

	_, err = NewTelegramProvider(TelegramConfig{Enabled: true, BotToken: "token"}, nil).Deliver(context.Background(), notification.Notification{})
	if err == nil || !strings.Contains(err.Error(), "chat_id") {
		t.Fatalf("error = %v, want chat_id error", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad chat", http.StatusBadRequest)
	}))
	defer server.Close()

	_, err = NewTelegramProvider(TelegramConfig{
		Enabled:    true,
		BotToken:   "token",
		APIBaseURL: server.URL,
	}, server.Client()).Deliver(context.Background(), notification.Notification{Recipient: "123", Title: "Hi"})
	if !jobs.IsPermanent(err) {
		t.Fatalf("error = %v, want permanent error", err)
	}
}

func TestNewConfiguredDispatcherRegistersExternalProviders(t *testing.T) {
	// Verify that BuildProvider can build various channels correctly
	_, err := BuildProvider(notification.ChannelEmail, `{"host":"smtp.mail.com","from":"noreply@test.com"}`, nil)
	if err != nil {
		t.Fatalf("build email: %v", err)
	}

	_, err = BuildProvider(notification.ChannelSMS, `{"url_template":"http://sms"}`, nil)
	if err != nil {
		t.Fatalf("build sms: %v", err)
	}

	_, err = BuildProvider(notification.ChannelWebhook, `{"url_template":"http://hook"}`, nil)
	if err != nil {
		t.Fatalf("build webhook: %v", err)
	}

	_, err = BuildProvider(notification.ChannelTelegram, `{"bot_token":"123"}`, nil)
	if err != nil {
		t.Fatalf("build telegram: %v", err)
	}

	_, err = BuildProvider(notification.ChannelSlack, `{"url_template":"http://slack"}`, nil)
	if err != nil {
		t.Fatalf("build slack: %v", err)
	}
}

func TestSMTPTransportSendsMessageOverPlainSMTP(t *testing.T) {
	server := newFakeSMTPServer(t)
	provider := NewSMTPProvider(EmailConfig{
		Enabled: true,
		Host:    server.host,
		Port:    server.port,
		From:    "noreply@example.com",
		TLSMode: "none",
		Timeout: time.Second,
	}, nil)

	_, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_smtp",
		Recipient: "alice@example.com",
		Channel:   notification.ChannelEmail,
		Title:     "SMTP test",
		Body:      "Plain SMTP body",
	})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}

	select {
	case message := <-server.messages:
		if !strings.Contains(message, "Plain SMTP body") {
			t.Fatalf("smtp message = %q, want body", message)
		}
	case <-time.After(time.Second):
		t.Fatal("smtp server did not receive message")
	}
}

func TestSMTPProviderRejectsInvalidConfiguration(t *testing.T) {
	record := notification.Notification{
		ID:        "notif_email",
		Recipient: "alice@example.com",
		Channel:   notification.ChannelEmail,
		Title:     "Hello",
		Body:      "World",
	}

	_, err := NewSMTPProvider(EmailConfig{Enabled: true, From: "noreply@example.com"}, &fakeSMTPTransport{}).Deliver(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "host") {
		t.Fatalf("error = %v, want host error", err)
	}

	_, err = NewSMTPProvider(EmailConfig{Enabled: true, Host: "smtp.example.com"}, &fakeSMTPTransport{}).Deliver(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "from") {
		t.Fatalf("error = %v, want from error", err)
	}

	record.Recipient = ""
	_, err = NewSMTPProvider(EmailConfig{Enabled: true, Host: "smtp.example.com", From: "noreply@example.com"}, &fakeSMTPTransport{}).Deliver(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "recipient") {
		t.Fatalf("error = %v, want recipient error", err)
	}
}

func TestSMTPProviderPropagatesTransportError(t *testing.T) {
	transportErr := errors.New("smtp down")
	_, err := NewSMTPProvider(EmailConfig{
		Enabled: true,
		Host:    "smtp.example.com",
		From:    "noreply@example.com",
	}, &fakeSMTPTransport{err: transportErr}).Deliver(context.Background(), notification.Notification{
		ID:        "notif_email",
		Recipient: "alice@example.com",
		Channel:   notification.ChannelEmail,
		Title:     "Hello",
		Body:      "World",
	})

	if !errors.Is(err, transportErr) {
		t.Fatalf("error = %v, want transport error", err)
	}
}

func TestHTTPProviderHandlesErrorsAndResponseIDFallbacks(t *testing.T) {
	_, err := NewHTTPProvider(HTTPProviderConfig{Name: "sms"}, nil).Deliver(context.Background(), notification.Notification{ID: "notif_sms"})
	if !errors.Is(err, ErrProviderDisabled) {
		t.Fatalf("error = %v, want ErrProviderDisabled", err)
	}

	_, err = NewHTTPProvider(HTTPProviderConfig{Name: "sms", Enabled: true}, nil).Deliver(context.Background(), notification.Notification{ID: "notif_sms"})
	if err == nil || !strings.Contains(err.Error(), "url template") {
		t.Fatalf("error = %v, want missing URL template", err)
	}

	failedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad provider request", http.StatusBadRequest)
	}))
	defer failedServer.Close()

	_, err = NewHTTPProvider(HTTPProviderConfig{
		Name:        "sms",
		Enabled:     true,
		URLTemplate: failedServer.URL,
	}, failedServer.Client()).Deliver(context.Background(), notification.Notification{ID: "notif_sms"})
	if err == nil || !strings.Contains(err.Error(), "status=400") {
		t.Fatalf("error = %v, want status failure", err)
	}

	headerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Message-ID", "header_123")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer headerServer.Close()

	got, err := NewHTTPProvider(HTTPProviderConfig{
		Name:             "sms",
		Enabled:          true,
		URLTemplate:      headerServer.URL,
		ResponseIDHeader: "X-Message-ID",
	}, headerServer.Client()).Deliver(context.Background(), notification.Notification{ID: "notif_sms"})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "header_123" {
		t.Fatalf("provider message ID = %q, want header_123", got.ProviderMessageID)
	}

	jsonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":42}}`))
	}))
	defer jsonServer.Close()

	got, err = NewHTTPProvider(HTTPProviderConfig{
		Name:                "sms",
		Enabled:             true,
		URLTemplate:         jsonServer.URL,
		ResponseIDJSONField: "data.id",
	}, jsonServer.Client()).Deliver(context.Background(), notification.Notification{ID: "notif_sms"})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "42" {
		t.Fatalf("provider message ID = %q, want 42", got.ProviderMessageID)
	}
}

func TestTemplateAndWebhookHelpers(t *testing.T) {
	rendered, err := renderTemplate("test", `{{ .TenantID }}:{{ json .Metadata }}`, notification.Notification{
		TenantID: "tenant-a",
		Metadata: notification.Metadata{
			"source": "test",
		},
	})
	if err != nil {
		t.Fatalf("renderTemplate returned error: %v", err)
	}
	if rendered != `tenant-a:{"source":"test"}` {
		t.Fatalf("rendered = %q", rendered)
	}

	if !hostAllowed("api.tenant.example.com", []string{"*.tenant.example.com"}) {
		t.Fatal("wildcard host should be allowed")
	}
	if hostAllowed("tenant.example.com", []string{"*.tenant.example.com"}) {
		t.Fatal("wildcard should require a subdomain")
	}
	if err := validateWebhookTarget("ftp://example.com/hook", WebhookConfig{}); !errors.Is(err, ErrWebhookTargetNotAllowed) {
		t.Fatalf("error = %v, want ErrWebhookTargetNotAllowed", err)
	}
	if err := validateWebhookTarget("https:///missing-host", WebhookConfig{}); !errors.Is(err, ErrWebhookTargetNotAllowed) {
		t.Fatalf("error = %v, want ErrWebhookTargetNotAllowed", err)
	}

	_, err = renderTemplate("bad", `{{ .Missing }}`, notification.Notification{})
	if err == nil {
		t.Fatal("missing template key should fail")
	}
}

func TestSMTPTransportEarlyErrors(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	err := smtpTransport{}.Send(canceled, EmailConfig{Host: "smtp.example.com", Port: 25, TLSMode: "none"}, "noreply@example.com", []string{"alice@example.com"}, []byte("body"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}

	err = smtpTransport{}.Send(context.Background(), EmailConfig{Host: "smtp.example.com", Port: 25, TLSMode: "invalid"}, "noreply@example.com", []string{"alice@example.com"}, []byte("body"))
	if err == nil || !strings.Contains(err.Error(), "unsupported tls mode") {
		t.Fatalf("error = %v, want unsupported tls mode", err)
	}

	if _, err := parseEmailAddress("not-an-address"); err == nil {
		t.Fatal("invalid email address should fail")
	}
}

func TestInAppProviderRequiresStore(t *testing.T) {
	_, err := NewInAppProvider(nil).Deliver(context.Background(), notification.Notification{
		ID:      "notif_inapp",
		Channel: notification.ChannelInApp,
	})

	if err == nil || !strings.Contains(err.Error(), "store") {
		t.Fatalf("error = %v, want store error", err)
	}
}

type fakeSMTPServer struct {
	host     string
	port     int
	messages chan string
	listener net.Listener
}

func newFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen smtp: %v", err)
	}
	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split smtp addr: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse smtp port: %v", err)
	}
	server := &fakeSMTPServer{
		host:     host,
		port:     port,
		messages: make(chan string, 1),
		listener: listener,
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handleFakeSMTPConnection(conn, server.messages)
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("fake smtp server did not stop")
		}
	})
	return server
}

func handleFakeSMTPConnection(conn net.Conn, messages chan<- string) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeLine := func(line string) {
		_, _ = fmt.Fprint(writer, line+"\r\n")
		_ = writer.Flush()
	}

	writeLine("220 test.smtp.local ESMTP")
	var message strings.Builder
	inData := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if inData {
			if line == "." {
				messages <- message.String()
				writeLine("250 queued")
				inData = false
				continue
			}
			message.WriteString(line)
			message.WriteByte('\n')
			continue
		}

		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			writeLine("250 test.smtp.local")
		case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
			writeLine("250 ok")
		case upper == "DATA":
			writeLine("354 end with dot")
			inData = true
		case upper == "QUIT":
			writeLine("221 bye")
			return
		default:
			writeLine("250 ok")
		}
	}
}
