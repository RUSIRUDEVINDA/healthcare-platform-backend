package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/auth-service/internal/model"
	"healthcare-platform/services/auth-service/internal/service"
	"healthcare-platform/pkg/logger"
)

// AuthHandler handles HTTP layer only
// It parses requests, calls the service, and formats responses
// NO business logic here — that lives in service/
type AuthHandler struct {
	authSvc *service.AuthService
	log     *logger.Logger
}

func NewAuthHandler(authSvc *service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, log: log}
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	// Public health check
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck) 

	// Auth endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/refresh", h.Refresh)

		// Internal endpoint — called only by API Gateway / other services
		// Not exposed publicly via Nginx
		auth.GET("/validate", h.ValidateToken)
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Registers patient or doctor, returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.RegisterRequest true "Register payload"
// @Success      201  {object} model.APIResponse
// @Failure      400  {object} model.APIResponse
// @Failure      409  {object} model.APIResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	tokenResp, err := h.authSvc.Register(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, model.ErrorResponse(err.Error()))
		default:
			h.log.Error("Register error", "error", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse("Registration failed. Please try again."))
		}
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse(tokenResp))
}

// Login godoc
// @Summary      Login
// @Description  Authenticates user and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.LoginRequest true "Login payload"
// @Success      200  {object} model.APIResponse
// @Failure      401  {object} model.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	tokenResp, err := h.authSvc.Login(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials),
			errors.Is(err, service.ErrAccountInactive):
			c.JSON(http.StatusUnauthorized, model.ErrorResponse(err.Error()))
		default:
			h.log.Error("Login error", "error", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse("Login failed. Please try again."))
		}
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(tokenResp))
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidates the refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.LogoutRequest true "Logout payload"
// @Success      200  {object} model.APIResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req model.LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	if err := h.authSvc.Logout(req.RefreshToken); err != nil {
		h.log.Error("Logout error", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Logout failed"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("Logged out successfully"))
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Issues new access + refresh tokens using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.RefreshRequest true "Refresh token payload"
// @Success      200  {object} model.APIResponse
// @Failure      401  {object} model.APIResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	tokenResp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken),
			errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, model.ErrorResponse(err.Error()))
		default:
			h.log.Error("Refresh error", "error", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse("Token refresh failed"))
		}
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(tokenResp))
}

// ValidateToken godoc
// @Summary      Validate JWT (internal)
// @Description  Called by API Gateway to validate Authorization header tokens
// @Tags         internal
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} model.ValidateTokenResponse
// @Failure      401  {object} model.ValidateTokenResponse
// @Router       /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, model.ValidateTokenResponse{Valid: false})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	resp, err := h.authSvc.ValidateToken(tokenStr)
	if err != nil || !resp.Valid {
		c.JSON(http.StatusUnauthorized, model.ValidateTokenResponse{Valid: false})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ──────────────────────────────────────────────
// Health Checks (used by Kubernetes probes)
// ──────────────────────────────────────────────

func (h *AuthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}

func (h *AuthHandler) ReadinessCheck(c *gin.Context) {
	// K8s readiness probe — returns 200 only when fully ready
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "auth-service",
	})
}

// ──────────────────────────────────────────────
// Private helpers
// ──────────────────────────────────────────────

func formatValidationError(err error) string {
	return "Invalid request: " + err.Error()
}
