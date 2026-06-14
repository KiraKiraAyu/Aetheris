package notification

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid notification request")
	ErrNotFound       = errors.New("notification not found")
)
