package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
)

// CORS allows browser clients to reach the gateway (same pattern as other services).
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

// Logger logs each request with latency.
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Info("gateway request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"client_ip", c.ClientIP(),
		)
	}
}

// InternalValidateOnly mirrors nginx: /api/auth/validate is only reachable from private networks.
func InternalValidateOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := stripQuery(c.Request.URL.Path)
		if path != "/api/auth/validate" {
			c.Next()
			return
		}

		host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
		if err != nil {
			host = strings.TrimSpace(c.Request.RemoteAddr)
		}
		ip := net.ParseIP(host)
		if ip == nil || !isAllowedInternal(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "this endpoint is only available from the internal network",
			})
			return
		}
		c.Next()
	}
}

func stripQuery(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		return p[:i]
	}
	return p
}

func isAllowedInternal(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	_, n10, _ := net.ParseCIDR("10.0.0.0/8")
	_, n172, _ := net.ParseCIDR("172.16.0.0/12")
	_, n192, _ := net.ParseCIDR("192.168.0.0/16")
	return n10.Contains(ip) || n172.Contains(ip) || n192.Contains(ip)
}
