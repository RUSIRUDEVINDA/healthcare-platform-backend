package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/telemedicine-service/internal/middleware"
	"healthcare-platform/services/telemedicine-service/internal/model"
	"healthcare-platform/services/telemedicine-service/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
	log *logger.Logger
}

func NewSessionHandler(svc *service.SessionService, log *logger.Logger) *SessionHandler {
	return &SessionHandler{svc: svc, log: log}
}

func (h *SessionHandler) RegisterRoutes(router *gin.Engine, auth gin.HandlerFunc) {
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api/v1")
	api.Use(auth)

	telemedicine := api.Group("/telemedicine")
	{
		telemedicine.POST("/appointments/:appointment_id/sessions", h.CreateSession)
		telemedicine.GET("/appointments/:appointment_id/sessions", h.ListSessions)
		telemedicine.GET("/sessions/:id", h.GetSession)
		telemedicine.GET("/sessions/:id/join", h.JoinSession)
		telemedicine.PATCH("/sessions/:id/start", h.StartSession)
		telemedicine.PATCH("/sessions/:id/end", h.EndSession)
	}
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can create sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.svc.CreateSession(callerID, callerRole, callerToken, strings.TrimSpace(c.Param("appointment_id")), &req)
	if err != nil {
		h.writeError(c, err, "failed to create telemedicine session")
		return
	}

	c.JSON(http.StatusCreated, session)
}

func (h *SessionHandler) ListSessions(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can view sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	sessions, err := h.svc.ListSessions(strings.TrimSpace(c.Param("appointment_id")), callerID, callerRole, callerToken)
	if err != nil {
		h.writeError(c, err, "failed to list sessions")
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func (h *SessionHandler) GetSession(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can view sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	session, err := h.svc.GetSession(strings.TrimSpace(c.Param("id")), callerID, callerRole, callerToken)
	if err != nil {
		h.writeError(c, err, "failed to load session")
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) JoinSession(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can join sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	info, err := h.svc.GetJoinInfo(strings.TrimSpace(c.Param("id")), callerID, callerRole, callerToken)
	if err != nil {
		h.writeError(c, err, "failed to generate join info")
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *SessionHandler) StartSession(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can update sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	session, err := h.svc.StartSession(strings.TrimSpace(c.Param("id")), callerID, callerRole, callerToken)
	if err != nil {
		h.writeError(c, err, "failed to start session")
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) EndSession(c *gin.Context) {
	callerID, ok := middleware.CallerID(c)
	if !ok || callerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	callerRole, ok := middleware.CallerRole(c)
	if !ok || (callerRole != "patient" && callerRole != "doctor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the assigned patient or doctor can update sessions"})
		return
	}
	callerToken, ok := middleware.CallerToken(c)
	if !ok || callerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	session, err := h.svc.EndSession(strings.TrimSpace(c.Param("id")), callerID, callerRole, callerToken)
	if err != nil {
		h.writeError(c, err, "failed to end session")
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "telemedicine-service", "status": "healthy"})
}

func (h *SessionHandler) writeError(c *gin.Context, err error, fallback string) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "missing caller identity"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	case strings.Contains(msg, "missing caller token"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	case strings.Contains(msg, "not authorized"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	case strings.Contains(msg, "forbidden: not your appointment"):
		c.JSON(http.StatusForbidden, gin.H{"error": msg})
	case strings.Contains(msg, "only the assigned patient or doctor"):
		c.JSON(http.StatusForbidden, gin.H{"error": msg})
	case strings.Contains(msg, "appointment is cancelled"):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case strings.Contains(msg, "appointment not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case strings.Contains(msg, "cannot be started") || strings.Contains(msg, "cannot be ended"):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
