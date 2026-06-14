package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type Provider interface {
	Deliver(context.Context, notification.Notification) (jobs.DeliveryResult, error)
}

type Dispatcher struct {
	repo       notification.DeliveryRepository
	inAppStore notification.InAppStore
}

func NewConfiguredDispatcher(repo notification.DeliveryRepository, inAppStore notification.InAppStore) *Dispatcher {
	return &Dispatcher{repo: repo, inAppStore: inAppStore}
}

func (d *Dispatcher) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if d == nil {
		return jobs.DeliveryResult{}, ErrUnsupportedChannel
	}

	cfg, err := d.repo.GetChannelConfig(ctx, record.TenantID, record.Channel)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return jobs.DeliveryResult{}, fmt.Errorf("channel %s is not configured for tenant %s", record.Channel, record.TenantID)
		}
		return jobs.DeliveryResult{}, err
	}

	if !cfg.Enabled {
		return jobs.DeliveryResult{}, fmt.Errorf("channel %s is disabled for tenant %s", record.Channel, record.TenantID)
	}

	provider, err := BuildProvider(record.Channel, cfg.Config, d.inAppStore)
	if err != nil {
		return jobs.DeliveryResult{}, err
	}

	return provider.Deliver(ctx, record)
}

// JSON configurations and conversions

type EmailJSONConfig struct {
	Host           string            `json:"host"`
	Port           int               `json:"port"`
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	From           string            `json:"from"`
	TLSMode        string            `json:"tls_mode"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	Headers        map[string]string `json:"headers"`
}

func (c EmailJSONConfig) ToConfig() EmailConfig {
	timeout := 10 * time.Second
	if c.TimeoutSeconds > 0 {
		timeout = time.Duration(c.TimeoutSeconds) * time.Second
	}
	return EmailConfig{
		Enabled:  true,
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		From:     c.From,
		TLSMode:  c.TLSMode,
		Timeout:  timeout,
		Headers:  c.Headers,
	}
}

type HTTPProviderJSONConfig struct {
	URLTemplate         string            `json:"url_template"`
	Method              string            `json:"method"`
	Headers             map[string]string `json:"headers"`
	BodyTemplate        string            `json:"body_template"`
	TimeoutSeconds      int               `json:"timeout_seconds"`
	SuccessStatusMin    int               `json:"success_status_min"`
	SuccessStatusMax    int               `json:"success_status_max"`
	ResponseIDHeader    string            `json:"response_id_header"`
	ResponseIDJSONField string            `json:"response_id_json_field"`
	SigningSecret       string            `json:"signing_secret"`
}

func (c HTTPProviderJSONConfig) ToConfig(name string) HTTPProviderConfig {
	timeout := 10 * time.Second
	if c.TimeoutSeconds > 0 {
		timeout = time.Duration(c.TimeoutSeconds) * time.Second
	}
	minStatus := c.SuccessStatusMin
	if minStatus == 0 {
		minStatus = 200
	}
	maxStatus := c.SuccessStatusMax
	if maxStatus == 0 {
		maxStatus = 299
	}
	method := c.Method
	if method == "" {
		method = "POST"
	}
	return HTTPProviderConfig{
		Name:                name,
		Enabled:             true,
		URLTemplate:         c.URLTemplate,
		Method:              method,
		Headers:             c.Headers,
		BodyTemplate:        c.BodyTemplate,
		Timeout:             timeout,
		SuccessStatusMin:    minStatus,
		SuccessStatusMax:    maxStatus,
		ResponseIDHeader:    c.ResponseIDHeader,
		ResponseIDJSONField: c.ResponseIDJSONField,
		SigningSecret:       c.SigningSecret,
	}
}

type WebhookJSONConfig struct {
	URLTemplate         string            `json:"url_template"`
	Method              string            `json:"method"`
	Headers             map[string]string `json:"headers"`
	BodyTemplate        string            `json:"body_template"`
	TimeoutSeconds      int               `json:"timeout_seconds"`
	SuccessStatusMin    int               `json:"success_status_min"`
	SuccessStatusMax    int               `json:"success_status_max"`
	ResponseIDHeader    string            `json:"response_id_header"`
	ResponseIDJSONField string            `json:"response_id_json_field"`
	AllowedHosts        []string          `json:"allowed_hosts"`
	AllowPrivateIPs     bool              `json:"allow_private_ips"`
	SigningSecret       string            `json:"signing_secret"`
}

func (c WebhookJSONConfig) ToConfig() WebhookConfig {
	timeout := 10 * time.Second
	if c.TimeoutSeconds > 0 {
		timeout = time.Duration(c.TimeoutSeconds) * time.Second
	}
	minStatus := c.SuccessStatusMin
	if minStatus == 0 {
		minStatus = 200
	}
	maxStatus := c.SuccessStatusMax
	if maxStatus == 0 {
		maxStatus = 299
	}
	method := c.Method
	if method == "" {
		method = "POST"
	}
	return WebhookConfig{
		Enabled:             true,
		URLTemplate:         c.URLTemplate,
		Method:              method,
		Headers:             c.Headers,
		BodyTemplate:        c.BodyTemplate,
		Timeout:             timeout,
		SuccessStatusMin:    minStatus,
		SuccessStatusMax:    maxStatus,
		ResponseIDHeader:    c.ResponseIDHeader,
		ResponseIDJSONField: c.ResponseIDJSONField,
		AllowedHosts:        c.AllowedHosts,
		AllowPrivateIPs:     c.AllowPrivateIPs,
		SigningSecret:       c.SigningSecret,
	}
}

type TelegramJSONConfig struct {
	BotToken       string            `json:"bot_token"`
	APIBaseURL     string            `json:"api_base_url"`
	ParseMode      string            `json:"parse_mode"`
	DisableLink    bool              `json:"disable_link"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	Headers        map[string]string `json:"headers"`
	BodyTemplate   string            `json:"body_template"`
}

func (c TelegramJSONConfig) ToConfig() TelegramConfig {
	timeout := 10 * time.Second
	if c.TimeoutSeconds > 0 {
		timeout = time.Duration(c.TimeoutSeconds) * time.Second
	}
	apiURL := c.APIBaseURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org"
	}
	return TelegramConfig{
		Enabled:      true,
		BotToken:     c.BotToken,
		APIBaseURL:   apiURL,
		ParseMode:    c.ParseMode,
		DisableLink:  c.DisableLink,
		Timeout:      timeout,
		Headers:      c.Headers,
		BodyTemplate: c.BodyTemplate,
	}
}

func BuildProvider(channel notification.Channel, configJSON string, inAppStore notification.InAppStore) (Provider, error) {
	switch channel {
	case notification.ChannelEmail:
		var c EmailJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode email config: %w", err)
		}
		cfg := c.ToConfig()
		if cfg.Host == "" || cfg.From == "" {
			return nil, fmt.Errorf("email provider: host and from address are required")
		}
		return NewSMTPProvider(cfg, nil), nil
	case notification.ChannelSMS:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode sms config: %w", err)
		}
		cfg := c.ToConfig("sms")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("sms provider: url_template is required")
		}
		return NewHTTPProvider(cfg, http.DefaultClient), nil
	case notification.ChannelWebhook:
		var c WebhookJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode webhook config: %w", err)
		}
		cfg := c.ToConfig()
		return NewWebhookProvider(cfg, http.DefaultClient), nil
	case notification.ChannelTelegram:
		var c TelegramJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode telegram config: %w", err)
		}
		cfg := c.ToConfig()
		if cfg.BotToken == "" {
			return nil, fmt.Errorf("telegram provider: bot_token is required")
		}
		return NewTelegramProvider(cfg, http.DefaultClient), nil
	case notification.ChannelSlack:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode slack config: %w", err)
		}
		cfg := c.ToConfig("slack")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("slack provider: webhook url_template is required")
		}
		return NewSlackProvider(cfg, http.DefaultClient), nil
	case notification.ChannelDiscord:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode discord config: %w", err)
		}
		cfg := c.ToConfig("discord")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("discord provider: webhook url_template is required")
		}
		return NewDiscordProvider(cfg, http.DefaultClient), nil
	case notification.ChannelFeishu:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode feishu config: %w", err)
		}
		cfg := c.ToConfig("feishu")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("feishu provider: webhook url_template is required")
		}
		return NewFeishuProvider(cfg, http.DefaultClient), nil
	case notification.ChannelDingTalk:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode dingtalk config: %w", err)
		}
		cfg := c.ToConfig("dingtalk")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("dingtalk provider: webhook url_template is required")
		}
		return NewDingTalkProvider(cfg, http.DefaultClient), nil
	case notification.ChannelWeCom:
		var c HTTPProviderJSONConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("decode wecom config: %w", err)
		}
		cfg := c.ToConfig("wecom")
		if cfg.URLTemplate == "" {
			return nil, fmt.Errorf("wecom provider: webhook url_template is required")
		}
		return NewWeComProvider(cfg, http.DefaultClient), nil
	case notification.ChannelInApp:
		if inAppStore == nil {
			return nil, fmt.Errorf("in-app provider: store is required")
		}
		return NewInAppProvider(inAppStore), nil
	default:
		return nil, fmt.Errorf("unsupported provider channel: %s", channel)
	}
}
