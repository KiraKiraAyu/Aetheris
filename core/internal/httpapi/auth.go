package httpapi

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
)

type contextKey string

const tenantContextKey contextKey = "tenant_id"

type Authenticator interface {
	Authenticate(*http.Request) (string, bool)
}

type StaticAPIKeyAuthenticator struct {
	keys map[string]string
}

func NewStaticAPIKeyAuthenticator(keys map[string]string) *StaticAPIKeyAuthenticator {
	clone := make(map[string]string, len(keys))
	for key, tenantID := range keys {
		key = strings.TrimSpace(key)
		tenantID = strings.TrimSpace(tenantID)
		if key != "" && tenantID != "" {
			clone[key] = tenantID
		}
	}
	return &StaticAPIKeyAuthenticator{keys: clone}
}

func (a *StaticAPIKeyAuthenticator) Authenticate(request *http.Request) (string, bool) {
	if a == nil || len(a.keys) == 0 {
		return "", false
	}
	token := apiKeyFromRequest(request)
	if token == "" {
		return "", false
	}
	for key, tenantID := range a.keys {
		if subtle.ConstantTimeCompare([]byte(token), []byte(key)) == 1 {
			return tenantID, true
		}
	}
	return "", false
}

func apiKeyFromRequest(request *http.Request) string {
	if value := strings.TrimSpace(request.Header.Get("X-API-Key")); value != "" {
		return value
	}
	value := strings.TrimSpace(request.Header.Get("Authorization"))
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func withTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenantID)
}

func tenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantContextKey).(string)
	return tenantID, ok && tenantID != ""
}
