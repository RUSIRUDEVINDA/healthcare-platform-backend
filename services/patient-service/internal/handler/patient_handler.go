package handler

import (
	"errors"
	"net/http"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/patient-service/internal/model"
	"healthcare-platform/services/patient-service/internal/service"

	"github.com/gin-gonic/gin"
	"healthcare-platform/pkg/jwt"
	"healthcare-platform/pkg/middleware"
)

type PatientHandler struct {
	svc *service.PatientService
	log *logger.Logger
}

func NewPatientHandler(svc *service.PatientService, log *logger.Logger) *PatientHandler {
	return &PatientHandler{svc: svc, log: log}
}

func (h *PatientHandler) RegisterRoutes(router *gin.Engine, jwtHelper *jwt.Helper) {
	// Public health check
	router.GET("/health", h.HealthCheck)

	patient := router.Group("/api/v1/patient")
	patient.Use(middleware.AuthRequired(jwtHelper))
	{
		patient.GET("/profile", h.GetProfile)
		patient.PUT("/profile", h.UpdateProfile)
		patient.PATCH("/profile", h.PatchProfile)
		patient.DELETE("/profile", h.DeleteProfile)
	}
}

func (h *PatientHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user context"})
		return
	}

	p, err := h.svc.GetProfile(userID.(string))
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user context"})
		return
	}

	var req model.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateProfile(userID.(string), &req); err != nil {
		if errors.Is(err, service.ErrPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient profile not found"})
			return
		}
		h.log.Error("Failed to update profile", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Patient profile updated successfully"})
}

func (h *PatientHandler) PatchProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user context"})
		return
	}

	var req model.PatchPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.PatchProfile(userID.(string), &req); err != nil {
		if errors.Is(err, service.ErrPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient profile not found"})
			return
		}
		h.log.Error("Failed to patch profile", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Patient profile patched successfully"})
}

func (h *PatientHandler) DeleteProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user context"})
		return
	}

	if err := h.svc.DeleteProfile(userID.(string)); err != nil {
		if errors.Is(err, service.ErrPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient profile not found"})
			return
		}
		h.log.Error("Failed to delete profile", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Patient profile deleted successfully"})
}

func (h *PatientHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "patient-service", "status": "healthy"})
}
