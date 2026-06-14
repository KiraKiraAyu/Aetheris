package delivery

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Email    EmailConfig
	SMS      HTTPProviderConfig
	Webhook  WebhookConfig
	InApp    InAppConfig
	Telegram TelegramConfig
	Slack    HTTPProviderConfig
	Discord  HTTPProviderConfig
	Feishu   HTTPProviderConfig
	DingTalk HTTPProviderConfig
	WeCom    HTTPProviderConfig
}

type EmailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLSMode  string
	Timeout  time.Duration
	Headers  map[string]string
}

type HTTPProviderConfig struct {
	Name                string
	Enabled             bool
	URLTemplate         string
	Method              string
	Headers             map[string]string
	BodyTemplate        string
	Timeout             time.Duration
	SuccessStatusMin    int
	SuccessStatusMax    int
	ResponseIDHeader    string
	ResponseIDJSONField string
	SigningSecret       string
}

type TelegramConfig struct {
	Enabled      bool
	BotToken     string
	APIBaseURL   string
	ParseMode    string
	DisableLink  bool
	Timeout      time.Duration
	Headers      map[string]string
	BodyTemplate string
}

type WebhookConfig struct {
	Enabled             bool
	URLTemplate         string
	Method              string
	Headers             map[string]string
	BodyTemplate        string
	Timeout             time.Duration
	SuccessStatusMin    int
	SuccessStatusMax    int
	ResponseIDHeader    string
	ResponseIDJSONField string
	AllowedHosts        []string
	AllowPrivateIPs     bool
	SigningSecret       string
}

type InAppConfig struct {
	Enabled bool
}

func (c Config) Validate() error {
	if c.Email.Enabled {
		if strings.TrimSpace(c.Email.Host) == "" {
			return fmt.Errorf("email provider: SMTP_HOST is required")
		}
		if strings.TrimSpace(c.Email.From) == "" {
			return fmt.Errorf("email provider: SMTP_FROM is required")
		}
	}
	if c.SMS.Enabled && strings.TrimSpace(c.SMS.URLTemplate) == "" {
		return fmt.Errorf("sms provider: SMS_HTTP_URL is required")
	}
	if c.Telegram.Enabled && strings.TrimSpace(c.Telegram.BotToken) == "" {
		return fmt.Errorf("telegram provider: TELEGRAM_BOT_TOKEN is required")
	}
	if c.Slack.Enabled && strings.TrimSpace(c.Slack.URLTemplate) == "" {
		return fmt.Errorf("slack provider: SLACK_WEBHOOK_URL is required")
	}
	if c.Discord.Enabled && strings.TrimSpace(c.Discord.URLTemplate) == "" {
		return fmt.Errorf("discord provider: DISCORD_WEBHOOK_URL is required")
	}
	if c.Feishu.Enabled && strings.TrimSpace(c.Feishu.URLTemplate) == "" {
		return fmt.Errorf("feishu provider: FEISHU_WEBHOOK_URL is required")
	}
	if c.DingTalk.Enabled && strings.TrimSpace(c.DingTalk.URLTemplate) == "" {
		return fmt.Errorf("dingtalk provider: DINGTALK_WEBHOOK_URL is required")
	}
	if c.WeCom.Enabled && strings.TrimSpace(c.WeCom.URLTemplate) == "" {
		return fmt.Errorf("wecom provider: WECOM_WEBHOOK_URL is required")
	}
	return nil
}
