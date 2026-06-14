package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aetheris/internal/notification"
)

func TestTelegramProviderSendsMessage(t *testing.T) {
	var body bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottelegram-token/sendMessage" {
			t.Fatalf("path = %s, want /bottelegram-token/sendMessage", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		_, _ = body.ReadFrom(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":42}}`))
	}))
	defer server.Close()

	provider := NewTelegramProvider(TelegramConfig{
		Enabled:     true,
		BotToken:    "telegram-token",
		APIBaseURL:  server.URL,
		ParseMode:   "MarkdownV2",
		Timeout:     time.Second,
		Headers:     map[string]string{"X-Test": "yes"},
		DisableLink: true,
	}, server.Client())

	got, err := provider.Deliver(context.Background(), notification.Notification{
		ID:        "notif_telegram",
		TenantID:  "tenant-a",
		Recipient: "123456",
		Channel:   notification.ChannelTelegram,
		Title:     "Deploy finished",
		Body:      "Build 123 is green.",
	})

	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if got.ProviderMessageID != "telegram:42" {
		t.Fatalf("provider message ID = %q, want telegram:42", got.ProviderMessageID)
	}
	var payload map[string]any
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		t.Fatalf("telegram payload is not JSON: %v", err)
	}
	if payload["chat_id"] != "123456" {
		t.Fatalf("chat_id = %#v, want 123456", payload["chat_id"])
	}
	if payload["text"] != "Deploy finished\nBuild 123 is green." {
		t.Fatalf("text = %#v", payload["text"])
	}
	if payload["parse_mode"] != "MarkdownV2" || payload["disable_web_page_preview"] != true {
		t.Fatalf("telegram options = %#v", payload)
	}
}

func TestChatWebhookPresetsSendExpectedPayloads(t *testing.T) {
	cases := []struct {
		name      string
		channel   notification.Channel
		provider  func(string, *http.Client) Provider
		wantField string
	}{
		{
			name:    "slack",
			channel: notification.ChannelSlack,
			provider: func(url string, client *http.Client) Provider {
				return NewSlackProvider(chatWebhookConfig(url), client)
			},
			wantField: "text",
		},
		{
			name:    "discord",
			channel: notification.ChannelDiscord,
			provider: func(url string, client *http.Client) Provider {
				return NewDiscordProvider(chatWebhookConfig(url), client)
			},
			wantField: "content",
		},
		{
			name:    "feishu",
			channel: notification.ChannelFeishu,
			provider: func(url string, client *http.Client) Provider {
				return NewFeishuProvider(chatWebhookConfig(url), client)
			},
			wantField: "content.text",
		},
		{
			name:    "dingtalk",
			channel: notification.ChannelDingTalk,
			provider: func(url string, client *http.Client) Provider {
				return NewDingTalkProvider(chatWebhookConfig(url), client)
			},
			wantField: "text.content",
		},
		{
			name:    "wecom",
			channel: notification.ChannelWeCom,
			provider: func(url string, client *http.Client) Provider {
				return NewWeComProvider(chatWebhookConfig(url), client)
			},
			wantField: "text.content",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = body.ReadFrom(r.Body)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			provider := tc.provider(server.URL, server.Client())
			got, err := provider.Deliver(context.Background(), notification.Notification{
				ID:        "notif_chat",
				TenantID:  "tenant-a",
				Recipient: "ignored-by-webhook",
				Channel:   tc.channel,
				Title:     "Alert",
				Body:      "CPU is high.",
			})

			if err != nil {
				t.Fatalf("Deliver returned error: %v", err)
			}
			if got.ProviderMessageID == "" {
				t.Fatal("provider message ID should be set")
			}
			payload := decodeJSONMap(t, body.Bytes())
			if value := nestedValue(payload, tc.wantField); value != "Alert\nCPU is high." {
				t.Fatalf("%s = %#v, want combined text; payload=%#v", tc.wantField, value, payload)
			}
		})
	}
}

func TestTextTemplateDataCombinesTitleAndBody(t *testing.T) {
	rendered, err := renderTemplate("text", `{{ .Text }}`, notification.Notification{
		Title: "Title only",
	})
	if err != nil {
		t.Fatalf("renderTemplate returned error: %v", err)
	}
	if rendered != "Title only" {
		t.Fatalf("rendered = %q, want Title only", rendered)
	}

	rendered, err = renderTemplate("text", `{{ .Text }}`, notification.Notification{
		Body: "Body only",
	})
	if err != nil {
		t.Fatalf("renderTemplate returned error: %v", err)
	}
	if rendered != "Body only" {
		t.Fatalf("rendered = %q, want Body only", rendered)
	}
}

func chatWebhookConfig(url string) HTTPProviderConfig {
	return HTTPProviderConfig{
		Enabled:     true,
		URLTemplate: url,
		Timeout:     time.Second,
	}
}

func decodeJSONMap(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode JSON: %v; payload=%s", err, payload)
	}
	return decoded
}

func nestedValue(payload map[string]any, path string) any {
	var current any = payload
	for _, part := range splitPath(path) {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[part]
	}
	return current
}

func splitPath(path string) []string {
	var parts []string
	for _, part := range bytes.Split([]byte(path), []byte(".")) {
		parts = append(parts, string(part))
	}
	return parts
}
