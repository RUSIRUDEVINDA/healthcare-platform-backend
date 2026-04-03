package handler

import (
	"net/http"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/patient-service/internal/model"
	"healthcare-platform/services/patient-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	svc *service.PatientService
	log *logger.Logger
}

func NewPatientHandler(svc *service.PatientService, log *logger.Logger) *PatientHandler {
	return &PatientHandler{svc: svc, log: log}
}

func (h *PatientHandler) RegisterRoutes(router *gin.Engine) {
	// Public health check
	router.GET("/health", h.HealthCheck)

	patient := router.Group("/api/v1/patient")
	{
		patient.GET("/profile", h.GetProfile)
		patient.PUT("/profile", h.UpdateProfile)
	}
}

func (h *PatientHandler) GetProfile(c *gin.Context) {
	// In a real app, userID would come from a JWT context (set by middleware)
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user context"})
		return
	}

	p, err := h.svc.GetProfile(userID)
	if err != nil {
		h.log.Error("Failed to fetch profile", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patient profile"})
		return
	}

	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient profile not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req model.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateProfile(userID, &req); err != nil {
		h.log.Error("Failed to update profile", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Patient profile updated successfully"})
}

func (h *PatientHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "patient-service", "status": "healthy"})
}
