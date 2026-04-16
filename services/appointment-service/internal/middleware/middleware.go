package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/jwt"
)

// JWTAuth validates a Bearer token and stores caller claims in the Gin context.
func JWTAuth(jwtHelper *jwt.Helper) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not your appointment"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtHelper.ParseAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not your appointment"})
			return
		}

		c.Set("caller_id", claims.UserID)
		c.Set("caller_email", claims.Email)
		c.Set("caller_role", claims.Role)
		c.Set("caller_first_name", claims.FirstName)
		c.Set("caller_last_name", claims.LastName)
		c.Set("caller_token", tokenStr)
		c.Next()
	}
}

// RequireRoles blocks callers whose JWT role is not in the allowed list.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("caller_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid caller role"})
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you do not have permission to access this resource"})
	}
}

func CallerID(c *gin.Context) (string, bool) {
	v, ok := c.Get("caller_id")
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

func CallerRole(c *gin.Context) (string, bool) {
	v, ok := c.Get("caller_role")
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

func CallerToken(c *gin.Context) (string, bool) {
	v, ok := c.Get("caller_token")
	if !ok {
		return "", false
	}
	token, ok := v.(string)
	return token, ok
}

func CallerFirstName(c *gin.Context) (string, bool) {
	v, ok := c.Get("caller_first_name")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func CallerLastName(c *gin.Context) (string, bool) {
	v, ok := c.Get("caller_last_name")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
