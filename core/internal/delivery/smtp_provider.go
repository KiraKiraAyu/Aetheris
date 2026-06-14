package delivery

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
	"time"

	"aetheris/internal/jobs"
	"aetheris/internal/notification"
)

type SMTPTransport interface {
	Send(context.Context, EmailConfig, string, []string, []byte) error
}

type SMTPProvider struct {
	config    EmailConfig
	transport SMTPTransport
}

func NewSMTPProvider(config EmailConfig, transport SMTPTransport) *SMTPProvider {
	if config.Port == 0 {
		config.Port = 587
	}
	if config.TLSMode == "" {
		config.TLSMode = "starttls"
	}
	config.TLSMode = strings.ToLower(config.TLSMode)
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Headers == nil {
		config.Headers = map[string]string{}
	}
	if transport == nil {
		transport = smtpTransport{}
	}
	return &SMTPProvider{
		config:    config,
		transport: transport,
	}
}

func (p *SMTPProvider) Deliver(ctx context.Context, record notification.Notification) (jobs.DeliveryResult, error) {
	if !p.config.Enabled {
		return jobs.DeliveryResult{}, ErrProviderDisabled
	}
	if p.config.Host == "" {
		return jobs.DeliveryResult{}, fmt.Errorf("smtp delivery: host is required")
	}
	if p.config.From == "" {
		return jobs.DeliveryResult{}, fmt.Errorf("smtp delivery: from is required")
	}
	recipients := parseRecipients(record.Recipient)
	if len(recipients) == 0 {
		return jobs.DeliveryResult{}, fmt.Errorf("smtp delivery: recipient is required")
	}

	message, err := buildEmailMessage(p.config, record, recipients)
	if err != nil {
		return jobs.DeliveryResult{}, err
	}
	if err := p.transport.Send(ctx, p.config, p.config.From, recipients, message); err != nil {
		return jobs.DeliveryResult{}, err
	}
	return jobs.DeliveryResult{ProviderMessageID: "smtp:" + record.ID}, nil
}

func parseRecipients(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})
	recipients := make([]string, 0, len(fields))
	for _, field := range fields {
		if recipient := strings.TrimSpace(field); recipient != "" {
			recipients = append(recipients, recipient)
		}
	}
	return recipients
}

func buildEmailMessage(config EmailConfig, record notification.Notification, recipients []string) ([]byte, error) {
	headers := []string{
		"From: " + sanitizeHeaderValue(config.From),
		"To: " + sanitizeHeaderValue(strings.Join(recipients, ", ")),
		"Subject: " + mime.QEncoding.Encode("utf-8", sanitizeHeaderValue(record.Title)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"X-Aetheris-Notification-ID: " + sanitizeHeaderValue(record.ID),
	}
	keys := make([]string, 0, len(config.Headers))
	for key := range config.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		headers = append(headers, sanitizeHeaderValue(key)+": "+sanitizeHeaderValue(config.Headers[key]))
	}

	var message bytes.Buffer
	for _, header := range headers {
		message.WriteString(header)
		message.WriteString("\r\n")
	}
	message.WriteString("\r\n")
	message.WriteString(record.Body)
	return message.Bytes(), nil
}

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

type smtpTransport struct{}

func (smtpTransport) Send(ctx context.Context, config EmailConfig, from string, recipients []string, message []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fromAddress, err := parseEmailAddress(from)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	dialer := net.Dialer{Timeout: config.Timeout}

	var conn net.Conn
	switch config.TLSMode {
	case "tls":
		conn, err = tls.DialWithDialer(&dialer, "tcp", addr, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: config.Host,
		})
	case "starttls", "none":
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	default:
		return fmt.Errorf("smtp delivery: unsupported tls mode %q", config.TLSMode)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if config.Timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(config.Timeout))
	}

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if config.TLSMode == "starttls" {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return fmt.Errorf("smtp delivery: server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: config.Host,
		}); err != nil {
			return err
		}
	}
	if config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", config.Username, config.Password, config.Host)); err != nil {
			return err
		}
	}
	if err := client.Mail(fromAddress); err != nil {
		return err
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func parseEmailAddress(value string) (string, error) {
	address, err := mail.ParseAddress(value)
	if err != nil {
		if strings.Contains(value, "@") {
			return value, nil
		}
		return "", err
	}
	return address.Address, nil
}
