package delivery

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type WebhookProvider struct {
	config WebhookConfig
	client *http.Client
}

func NewWebhookProvider(config WebhookConfig, client *http.Client) *WebhookProvider {
	if config.URLTemplate == "" {
		config.URLTemplate = "{{ .Recipient }}"
	}
	if config.Method == "" {
		config.Method = http.MethodPost
	}
	if config.BodyTemplate == "" {
		config.BodyTemplate = defaultHTTPBodyTemplate
	}
	if config.SuccessStatusMin == 0 {
		config.SuccessStatusMin = 200
	}
	if config.SuccessStatusMax == 0 {
		config.SuccessStatusMax = 299
	}
	if config.Headers == nil {
		config.Headers = map[string]string{}
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &WebhookProvider{
		config: config,
		client: client,
	}
}

func (p *WebhookProvider) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if !p.config.Enabled {
		return jobs.DeliveryResult{}, ErrProviderDisabled
	}
	endpoint, err := renderTemplate("webhook.url", p.config.URLTemplate, record)
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("render webhook url: %w", err)
	}
	if err := validateWebhookTarget(endpoint, p.config); err != nil {
		return jobs.DeliveryResult{}, err
	}

	return NewHTTPProvider(HTTPProviderConfig{
		Name:                "webhook",
		Enabled:             true,
		URLTemplate:         endpoint,
		Method:              p.config.Method,
		Headers:             p.config.Headers,
		BodyTemplate:        p.config.BodyTemplate,
		Timeout:             p.config.Timeout,
		SuccessStatusMin:    p.config.SuccessStatusMin,
		SuccessStatusMax:    p.config.SuccessStatusMax,
		ResponseIDHeader:    p.config.ResponseIDHeader,
		ResponseIDJSONField: p.config.ResponseIDJSONField,
		SigningSecret:       p.config.SigningSecret,
	}, p.client).Deliver(ctx, record)
}

func validateWebhookTarget(rawURL string, config WebhookConfig) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: invalid url", ErrWebhookTargetNotAllowed)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: unsupported scheme %q", ErrWebhookTargetNotAllowed, parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("%w: missing host", ErrWebhookTargetNotAllowed)
	}
	if len(config.AllowedHosts) > 0 && !hostAllowed(host, config.AllowedHosts) {
		return fmt.Errorf("%w: host %q is not in allowlist", ErrWebhookTargetNotAllowed, host)
	}
	if ip := net.ParseIP(host); ip != nil && !config.AllowPrivateIPs && isPrivateWebhookIP(ip) {
		return fmt.Errorf("%w: private ip %q", ErrWebhookTargetNotAllowed, host)
	}
	return nil
}

func hostAllowed(host string, allowedHosts []string) bool {
	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSpace(allowed))
		if allowed == "" {
			continue
		}
		if allowed == "*" || allowed == host {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(host, suffix) {
				return true
			}
		}
	}
	return false
}

func isPrivateWebhookIP(ip net.IP) bool {
	return ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}
