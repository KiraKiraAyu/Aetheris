package delivery

import "errors"

var (
	ErrProviderDisabled        = errors.New("delivery provider disabled")
	ErrUnsupportedChannel      = errors.New("unsupported notification channel")
	ErrWebhookTargetNotAllowed = errors.New("webhook target not allowed")
)
