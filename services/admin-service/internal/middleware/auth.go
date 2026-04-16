package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/jwt"
	"healthcare-platform/services/admin-service/internal/model"
)

const (
	ContextUserID = "user_id"
	ContextEmail  = "email"
	ContextRole   = "role"
)

func RequireAuth(jwtHelper *jwt.Helper) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("authorization header required"))
			return
		}

		claims, err := jwtHelper.ParseAccessToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("invalid or expired token"))
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(ContextRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("no role found in context"))
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse("role type assertion failed"))
			return
		}

		if role != string(model.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse("insufficient permissions"))
			return
		}

		c.Next()
	}
}
