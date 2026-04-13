package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/ai-symptom-service/internal/middleware"
	"healthcare-platform/services/ai-symptom-service/internal/model"
	"healthcare-platform/services/ai-symptom-service/internal/repository"
	"healthcare-platform/services/ai-symptom-service/internal/service"
)

const defaultHistoryLimit = 20
const maxHistoryLimit = 100

// SymptomHandler exposes HTTP routes for the AI symptom checker.
type SymptomHandler struct {
	svc      *service.SymptomService
	chatRepo *repository.ChatRepository
	log      *logger.Logger
}

func NewSymptomHandler(svc *service.SymptomService, chatRepo *repository.ChatRepository, log *logger.Logger) *SymptomHandler {
	return &SymptomHandler{svc: svc, chatRepo: chatRepo, log: log}
}

// RegisterRoutes wires public health checks and JWT-protected symptom analysis.
func (h *SymptomHandler) RegisterRoutes(router *gin.Engine, authClient *http.Client, authBaseURL string) {
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	protected := router.Group("/symptoms")
	protected.Use(middleware.RequireAuthViaAuthService(authClient, authBaseURL))
	protected.Use(middleware.RequireRole("patient", "doctor", "admin"))
	{
		protected.POST("/check", h.Check)
		protected.GET("/history", h.ListHistory)
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

	if h.chatRepo != nil {
		userID, ok := c.Get(middleware.ContextUserID)
		if ok {
			if uid, ok := userID.(string); ok && uid != "" {
				if _, err := h.chatRepo.Save(c.Request.Context(), uid, &req, out); err != nil {
					h.log.Error("Failed to save symptom chat history", "error", err)
				}
			}
		}
	}

	c.JSON(http.StatusOK, model.SuccessResponse(out))
}

func (h *SymptomHandler) ListHistory(c *gin.Context) {
	if h.chatRepo == nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse("Chat history is not configured (DATABASE_URL missing)"))
		return
	}
	userIDVal, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Not authenticated"))
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Invalid user context"))
		return
	}

	limit := parseQueryInt(c, "limit", defaultHistoryLimit, maxHistoryLimit)
	offset := parseQueryOffset(c, "offset")

	items, err := h.chatRepo.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.log.Error("List symptom chat history failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to load chat history"))
		return
	}
	if items == nil {
		items = []model.SymptomChatHistoryItem{}
	}
	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{"items": items}))
}

func parseQueryInt(c *gin.Context, name string, defaultVal, maxVal int) int {
	s := c.Query(name)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return defaultVal
	}
	if n > maxVal {
		return maxVal
	}
	return n
}

func parseQueryOffset(c *gin.Context, name string) int {
	s := c.Query(name)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (h *SymptomHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "ai-symptom-service",
	})
}

func (h *SymptomHandler) ReadinessCheck(c *gin.Context) {
	if h.chatRepo != nil {
		if err := h.chatRepo.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"service": "ai-symptom-service",
				"error":   "database unavailable",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "ai-symptom-service",
	})
}
