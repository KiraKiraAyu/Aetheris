package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"aetheris/internal/notification"
)

type NotificationService interface {
	Create(context.Context, notification.CreateRequest) (notification.Notification, error)
	List(context.Context, notification.NotificationQuery) ([]notification.Notification, error)
	Get(context.Context, string, string) (notification.Notification, error)
	ListDeliveryAttempts(context.Context, string, string) ([]notification.DeliveryAttempt, error)
	ListInApp(context.Context, notification.InAppQuery) ([]notification.InAppMessage, error)
	MarkInAppRead(context.Context, string, string, string) error
	ListTemplates(context.Context, notification.TemplateQuery) ([]notification.NotificationTemplate, error)
	SaveTemplate(context.Context, notification.NotificationTemplate) (notification.NotificationTemplate, error)
	DeleteTemplate(context.Context, string, string) error
	ListChannelConfigs(ctx context.Context, tenantID string) ([]notification.ChannelConfig, error)
	SaveChannelConfig(ctx context.Context, cfg notification.ChannelConfig) (notification.ChannelConfig, error)
}

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

type Options struct {
	Authenticator Authenticator
	RateLimiter   RateLimiter
	AllowedOrigins []string
	MaxBodyBytes  int64
}

func RegisterRoutes(router gin.IRouter, service NotificationService) {
	RegisterRoutesWithOptions(router, service, Options{})
}

func RegisterRoutesWithOptions(router gin.IRouter, service NotificationService, options Options) {
	apiGroup := router.Group("")
	if len(options.AllowedOrigins) > 0 {
		apiGroup.Use(corsMiddleware(options.AllowedOrigins))
	}

	apiGroup.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	routes := apiGroup.Group("")
	if options.Authenticator != nil {
		routes.Use(authMiddleware(options.Authenticator))
	}
	if options.RateLimiter != nil {
		routes.Use(rateLimitMiddleware(options.RateLimiter))
	}
	if options.MaxBodyBytes > 0 {
		routes.Use(maxBodyMiddleware(options.MaxBodyBytes))
	}

	routes.POST("/send", createNotification(service))
	routes.OPTIONS("/*path", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})
	routes.GET("/notifications", listNotifications(service))
	routes.GET("/notifications/:id", getNotification(service))
	routes.GET("/notifications/:id/attempts", listDeliveryAttempts(service))
	routes.GET("/in-app/messages", listInAppMessages(service))
	routes.POST("/in-app/messages/:id/read", markInAppRead(service))
	routes.GET("/templates", listTemplates(service))
	routes.POST("/templates", saveTemplate(service))
	routes.DELETE("/templates/:id", deleteTemplate(service))
	routes.GET("/channels", listChannelConfigs(service))
	routes.POST("/channels", saveChannelConfig(service))
}

func createNotification(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request notification.CreateRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "malformed JSON request"})
			return
		}
		if !applyTenantScope(ctx, &request.TenantID) {
			return
		}

		created, err := service.Create(ctx.Request.Context(), request)
		if err != nil {
			if errors.Is(err, notification.ErrInvalidRequest) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "create notification"})
			return
		}

		ctx.JSON(http.StatusAccepted, created)
	}
}

func listNotifications(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := notification.NotificationQuery{
			TenantID:  tenantForQuery(ctx),
			Recipient: ctx.Query("recipient"),
			Channel:   notification.Channel(ctx.Query("channel")),
			Status:    notification.Status(ctx.Query("status")),
			Limit:     parseLimit(ctx.Query("limit")),
		}
		if query.TenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		notifications, err := service.List(ctx.Request.Context(), query)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, notifications)
	}
}

func getNotification(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		record, err := service.Get(ctx.Request.Context(), tenantID, ctx.Param("id"))
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, record)
	}
}

func listDeliveryAttempts(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		attempts, err := service.ListDeliveryAttempts(ctx.Request.Context(), tenantID, ctx.Param("id"))
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, attempts)
	}
}

func listInAppMessages(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := notification.InAppQuery{
			TenantID:   tenantForQuery(ctx),
			UserID:     ctx.Query("user_id"),
			UnreadOnly: parseBool(ctx.Query("unread")),
			Limit:      parseLimit(ctx.Query("limit")),
		}
		if query.TenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		messages, err := service.ListInApp(ctx.Request.Context(), query)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, messages)
	}
}

func markInAppRead(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		if err := service.MarkInAppRead(ctx.Request.Context(), tenantID, ctx.Param("id"), ctx.Query("user_id")); err != nil {
			writeError(ctx, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listTemplates(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := notification.TemplateQuery{
			TenantID: tenantForQuery(ctx),
			Channel:  notification.Channel(ctx.Query("channel")),
			Key:      ctx.Query("key"),
			Limit:    parseLimit(ctx.Query("limit")),
		}
		if query.TenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		templates, err := service.ListTemplates(ctx.Request.Context(), query)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, templates)
	}
}

func saveTemplate(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var tpl notification.NotificationTemplate
		if err := ctx.ShouldBindJSON(&tpl); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "malformed JSON request"})
			return
		}
		if !applyTenantScope(ctx, &tpl.TenantID) {
			return
		}
		saved, err := service.SaveTemplate(ctx.Request.Context(), tpl)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, saved)
	}
}

func deleteTemplate(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		if err := service.DeleteTemplate(ctx.Request.Context(), tenantID, ctx.Param("id")); err != nil {
			writeError(ctx, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			headers := ctx.Writer.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Credentials", "true")
			headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			headers.Add("Vary", "Origin")
			if ctx.Request.Method == http.MethodOptions {
				ctx.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		ctx.Next()
	}
}

func authMiddleware(authenticator Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID, ok := authenticator.Authenticate(ctx.Request)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		ctx.Request = ctx.Request.WithContext(withTenant(ctx.Request.Context(), tenantID))
		ctx.Next()
	}
}

func rateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			tenantID = "anonymous"
		}
		allowed, err := limiter.Allow(ctx.Request.Context(), tenantID+":"+ctx.FullPath())
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limit"})
			return
		}
		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		ctx.Next()
	}
}

func maxBodyMiddleware(limit int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)
		}
		ctx.Next()
	}
}

func applyTenantScope(ctx *gin.Context, tenantID *string) bool {
	authTenant, ok := tenantFromContext(ctx.Request.Context())
	if !ok {
		return true
	}
	if *tenantID == "" {
		*tenantID = authTenant
		return true
	}
	if *tenantID != authTenant {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "tenant mismatch"})
		return false
	}
	return true
}

func tenantForQuery(ctx *gin.Context) string {
	if tenantID, ok := tenantFromContext(ctx.Request.Context()); ok {
		return tenantID
	}
	return ctx.Query("tenant_id")
}

func parseLimit(value string) int {
	if value == "" {
		return 0
	}
	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return limit
}

func parseBool(value string) bool {
	parsed, _ := strconv.ParseBool(value)
	return parsed
}

func listChannelConfigs(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantID := tenantForQuery(ctx)
		if tenantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
		configs, err := service.ListChannelConfigs(ctx.Request.Context(), tenantID)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, configs)
	}
}

func saveChannelConfig(service NotificationService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg notification.ChannelConfig
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "malformed JSON request"})
			return
		}
		if !applyTenantScope(ctx, &cfg.TenantID) {
			return
		}
		saved, err := service.SaveChannelConfig(ctx.Request.Context(), cfg)
		if err != nil {
			writeError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, saved)
	}
}

func writeError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, notification.ErrInvalidRequest):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, notification.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "request failed"})
	}
}
