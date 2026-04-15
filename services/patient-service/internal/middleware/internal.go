package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalAPIKey allows service-to-service calls (e.g. appointment-service fetching patient display names).
func InternalAPIKey(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "internal API not configured"})
			return
		}
		if c.GetHeader("X-Internal-Api-Key") != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
