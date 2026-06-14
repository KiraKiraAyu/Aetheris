package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type TelegramProvider struct {
	config TelegramConfig
	client *http.Client
}

func NewTelegramProvider(config TelegramConfig, client *http.Client) *TelegramProvider {
	if config.APIBaseURL == "" {
		config.APIBaseURL = "https://api.telegram.org"
	}
	config.APIBaseURL = strings.TrimRight(config.APIBaseURL, "/")
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Headers == nil {
		config.Headers = map[string]string{}
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &TelegramProvider{config: config, client: client}
}

func (p *TelegramProvider) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if !p.config.Enabled {
		return jobs.DeliveryResult{}, ErrProviderDisabled
	}
	if p.config.BotToken == "" {
		return jobs.DeliveryResult{}, fmt.Errorf("telegram delivery: bot token is required")
	}
	if strings.TrimSpace(record.Recipient) == "" {
		return jobs.DeliveryResult{}, fmt.Errorf("telegram delivery: recipient chat_id is required")
	}

	payload, err := p.telegramPayload(record)
	if err != nil {
		return jobs.DeliveryResult{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	endpoint := p.config.APIBaseURL + "/bot" + p.config.BotToken + "/sendMessage"
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range p.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("send telegram request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("read telegram response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("telegram delivery failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return jobs.DeliveryResult{}, jobs.Permanent(err)
		}
		return jobs.DeliveryResult{}, err
	}

	messageID := extractJSONField(responseBody, "result.message_id")
	if messageID == "" {
		messageID = fmt.Sprint(resp.StatusCode)
	}
	return jobs.DeliveryResult{ProviderMessageID: "telegram:" + messageID}, nil
}

func (p *TelegramProvider) telegramPayload(record notification.Notification) ([]byte, error) {
	if p.config.BodyTemplate != "" {
		rendered, err := renderTemplate("telegram.body", p.config.BodyTemplate, record)
		if err != nil {
			return nil, fmt.Errorf("render telegram body: %w", err)
		}
		return []byte(rendered), nil
	}

	payload := map[string]any{
		"chat_id": record.Recipient,
		"text":    notificationText(record),
	}
	if p.config.ParseMode != "" {
		payload["parse_mode"] = p.config.ParseMode
	}
	if p.config.DisableLink {
		payload["disable_web_page_preview"] = true
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}
