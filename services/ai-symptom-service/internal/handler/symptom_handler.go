package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/ai-symptom-service/internal/middleware"
	"healthcare-platform/services/ai-symptom-service/internal/model"
	"healthcare-platform/services/ai-symptom-service/internal/service"
)

// SymptomHandler exposes HTTP routes for the AI symptom checker.
type SymptomHandler struct {
	svc *service.SymptomService
	log *logger.Logger
}

func NewSymptomHandler(svc *service.SymptomService, log *logger.Logger) *SymptomHandler {
	return &SymptomHandler{svc: svc, log: log}
}

// RegisterRoutes wires public health checks and JWT-protected symptom analysis (patient only).
func (h *SymptomHandler) RegisterRoutes(router *gin.Engine, authClient *http.Client, authBaseURL string) {
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	protected := router.Group("/symptoms")
	protected.Use(middleware.RequireAuthViaAuthService(authClient, authBaseURL))
	protected.Use(middleware.RequireRole("patient"))
	{
		protected.POST("/check", h.Check)
	}
}

func (h *SymptomHandler) Check(c *gin.Context) {
	var req model.SymptomCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	out, err := h.svc.Check(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrLLMQuotaOrRateLimit) {
			c.JSON(http.StatusTooManyRequests, model.ErrorResponse(
				"AI quota or rate limit reached. Wait and retry, set GEMINI_MODEL to a model with free-tier quota (e.g. gemini-2.5-flash), or enable billing. See https://ai.google.dev/gemini-api/docs/rate-limits",
			))
			return
		}
		if errors.Is(err, service.ErrLLMOutputInvalid) {
			c.JSON(http.StatusBadGateway, model.ErrorResponse("Symptom analysis temporarily unavailable; please try again"))
			return
		}
		h.log.Error("Symptom check failed", "error", err)
		c.JSON(http.StatusBadGateway, model.ErrorResponse("Failed to analyze symptoms"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(out))
}

func (h *SymptomHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "ai-symptom-service",
	})
}

func (h *SymptomHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "ai-symptom-service",
	})
}
