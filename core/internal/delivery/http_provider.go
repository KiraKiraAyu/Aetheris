package delivery

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

const defaultHTTPBodyTemplate = `{"id":{{ quote .ID }},"tenant_id":{{ quote .TenantID }},"recipient":{{ quote .Recipient }},"channel":{{ quote .Channel }},"template_key":{{ quote .TemplateKey }},"title":{{ quote .Title }},"body":{{ quote .Body }},"group_key":{{ quote .GroupKey }},"aggregate_count":{{ .AggregateCount }},"metadata":{{ json .Metadata }}}`

type HTTPProvider struct {
	config HTTPProviderConfig
	client *http.Client
}

func NewHTTPProvider(config HTTPProviderConfig, client *http.Client) *HTTPProvider {
	config = normalizeHTTPProviderConfig(config)
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPProvider{
		config: config,
		client: client,
	}
}

func (p *HTTPProvider) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if !p.config.Enabled {
		return jobs.DeliveryResult{}, ErrProviderDisabled
	}
	if p.config.URLTemplate == "" {
		return jobs.DeliveryResult{}, fmt.Errorf("http delivery %s: url template is required", p.config.Name)
	}

	endpoint, err := renderTemplate(p.config.Name+".url", p.config.URLTemplate, record)
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("render %s url: %w", p.config.Name, err)
	}
	bodyText, err := renderTemplate(p.config.Name+".body", p.config.BodyTemplate, record)
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("render %s body: %w", p.config.Name, err)
	}
	body := []byte(bodyText)

	requestCtx := ctx
	cancel := func() {}
	if p.config.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, p.config.Timeout)
	}
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, p.config.Method, endpoint, bytes.NewReader(body))
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("build %s request: %w", p.config.Name, err)
	}
	if bodyText != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, valueTemplate := range p.config.Headers {
		value, err := renderTemplate(p.config.Name+".header."+key, valueTemplate, record)
		if err != nil {
			return jobs.DeliveryResult{}, fmt.Errorf("render %s header %s: %w", p.config.Name, key, err)
		}
		req.Header.Set(key, value)
	}
	if p.config.SigningSecret != "" {
		req.Header.Set("X-Aetheris-Signature", signPayload(p.config.SigningSecret, body))
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("send %s request: %w", p.config.Name, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return jobs.DeliveryResult{}, fmt.Errorf("read %s response: %w", p.config.Name, err)
	}
	if resp.StatusCode < p.config.SuccessStatusMin || resp.StatusCode > p.config.SuccessStatusMax {
		err := fmt.Errorf("%s delivery failed: status=%d body=%s", p.config.Name, resp.StatusCode, strings.TrimSpace(string(responseBody)))
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return jobs.DeliveryResult{}, jobs.Permanent(err)
		}
		return jobs.DeliveryResult{}, err
	}

	return jobs.DeliveryResult{
		ProviderMessageID: extractResponseID(p.config, resp, responseBody),
	}, nil
}

func normalizeHTTPProviderConfig(config HTTPProviderConfig) HTTPProviderConfig {
	if config.Name == "" {
		config.Name = "http"
	}
	if config.Method == "" {
		config.Method = http.MethodPost
	}
	config.Method = strings.ToUpper(config.Method)
	if config.BodyTemplate == "" {
		config.BodyTemplate = defaultHTTPBodyTemplate
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
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
	return config
}

func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func extractResponseID(config HTTPProviderConfig, resp *http.Response, body []byte) string {
	if config.ResponseIDHeader != "" {
		if value := resp.Header.Get(config.ResponseIDHeader); value != "" {
			return value
		}
	}
	if config.ResponseIDJSONField != "" && len(body) > 0 {
		if value := extractJSONField(body, config.ResponseIDJSONField); value != "" {
			return value
		}
	}
	return fmt.Sprintf("%s:%d", config.Name, resp.StatusCode)
}

func extractJSONField(body []byte, field string) string {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return ""
	}
	for _, part := range strings.Split(field, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		value, ok = object[part]
		if !ok {
			return ""
		}
	}
	switch typed := value.(type) {
	case string:
		return typed
	case float64, bool:
		return fmt.Sprint(typed)
	default:
		return ""
	}
}
