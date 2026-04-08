package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/doctor-service/internal/model"
	"healthcare-platform/pkg/logger"
)

// Context keys for values forwarded from auth-service validation.
const (
	ContextUserID = "user_id"
	ContextEmail  = "email"
	ContextRole   = "role"
)

// CORS allows browser clients to call the API (same pattern as auth-service).
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Logger logs each request with latency (same pattern as auth-service).
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"client_ip", c.ClientIP(),
		)
	}
}

// RequireAuthViaAuthService validates the Bearer token by calling auth-service GET /auth/validate.
func RequireAuthViaAuthService(client *http.Client, authBaseURL string) gin.HandlerFunc {
	base := strings.TrimRight(authBaseURL, "/")
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Authorization header required"))
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Authorization header must use Bearer scheme"))
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, base+"/auth/validate", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse("failed to build auth request"))
			return
		}
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, model.ErrorResponse("auth service unavailable"))
			return
		}
		defer resp.Body.Close()

		var body model.ValidateTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("invalid auth response"))
			return
		}

		if resp.StatusCode != http.StatusOK || !body.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("invalid or expired token"))
			return
		}

		c.Set(ContextUserID, body.UserID)
		c.Set(ContextEmail, body.Email)
		c.Set(ContextRole, body.Role)

		c.Next()
	}
}

// RequireRole restricts access to the given roles (use after RequireAuthViaAuthService).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("No role found in context"))
			return
		}
		roleStr, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse("Role type assertion failed"))
			return
		}

		for _, allowed := range allowedRoles {
			if roleStr == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse("Insufficient permissions"))
	}
}
