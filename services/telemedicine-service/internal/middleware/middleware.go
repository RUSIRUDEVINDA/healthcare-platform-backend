package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/jwt"
)

func JWTAuth(jwtHelper *jwt.Helper) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtHelper.ParseAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
			return
		}

		c.Set("caller_id", claims.UserID)
		c.Set("caller_email", claims.Email)
		c.Set("caller_role", claims.Role)
		c.Set("caller_token", tokenStr)
		c.Next()
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
