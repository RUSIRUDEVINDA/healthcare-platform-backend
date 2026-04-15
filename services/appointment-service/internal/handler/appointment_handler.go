package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/appointment-service/internal/middleware"
	"healthcare-platform/services/appointment-service/internal/model"
	"healthcare-platform/services/appointment-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	svc *service.AppointmentService
	log *logger.Logger
}

func NewAppointmentHandler(svc *service.AppointmentService, log *logger.Logger) *AppointmentHandler {
	return &AppointmentHandler{svc: svc, log: log}
}

func (h *AppointmentHandler) RegisterRoutes(router *gin.Engine, auth gin.HandlerFunc) {
	// Public health check
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api/v1")
	api.Use(auth)

	appointment := api.Group("/appointments")
	{
		appointment.POST("", middleware.RequireRoles("patient"), h.Book)
		appointment.GET("", middleware.RequireRoles("patient", "doctor", "admin"), h.ListAppointments)
		appointment.GET("/doctor/:doctor_id", middleware.RequireRoles("patient", "doctor", "admin"), h.GetSlots)
		appointment.GET("/:id", middleware.RequireRoles("patient", "doctor", "admin"), h.GetStatus)
		appointment.PUT("/:id/cancel", middleware.RequireRoles("patient", "doctor", "admin"), h.Cancel)
		appointment.PATCH("/:id/status", middleware.RequireRoles("doctor", "admin"), h.UpdateStatus)
	}

	slot := api.Group("/slots")
	{
		slot.POST("", middleware.RequireRoles("doctor", "admin"), h.CreateSlot)
		slot.GET("", middleware.RequireRoles("doctor", "admin"), h.ListSlots)
		slot.PUT("/:id", middleware.RequireRoles("doctor", "admin"), h.UpdateSlot)
		slot.DELETE("/:id", middleware.RequireRoles("doctor", "admin"), h.DeleteSlot)
	}
}

func (h *AppointmentHandler) Book(c *gin.Context) {
	userID, ok := middleware.CallerID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}

	role, ok := middleware.CallerRole(c)
	if !ok || role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only patients can book appointments"})
		return
	}

	token, ok := middleware.CallerToken(c)
	if !ok || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	var req model.BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.SlotID) != "" {
		if _, err := uuid.Parse(strings.TrimSpace(req.SlotID)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slot_id must be a valid UUID"})
			return
		}
	}

	appt, err := h.svc.BookAppointment(userID, role, token, &req)
	if err != nil {
		if strings.Contains(err.Error(), "only patients can book appointments") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "slot is already booked") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "no available slot found for requested time") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "pay later is not allowed") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "scheduled time must be in the future") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "invalid consultation mode") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "doctor channeling fee is not configured") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "doctor profile not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "doctor service unavailable") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "slot not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.log.Error("Failed to book appointment", "user_id", userID, "role", role, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to book appointment"})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

func (h *AppointmentHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	userID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)

	appt, err := h.svc.GetStatus(id, userID, role)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden: not your appointment") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
		return
	}
	if appt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		return
	}
	c.JSON(http.StatusOK, appt)
}

func (h *AppointmentHandler) ListAppointments(c *gin.Context) {
	userID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)
	appointments, err := h.svc.ListAppointments(userID, role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.CallerID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	role, ok := middleware.CallerRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
		return
	}

	if err := h.svc.CancelAppointment(id, userID, role); err != nil {
		if strings.Contains(err.Error(), "forbidden: not your appointment") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "appointment cancelled"})
}

func (h *AppointmentHandler) GetSlots(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		status = strings.TrimSpace(c.Query("filter"))
	}

	slots, err := h.svc.GetDoctorSlots(doctorID, status)
	if err != nil {
		if strings.Contains(err.Error(), "invalid slot status filter") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get slots"})
		return
	}
	c.JSON(http.StatusOK, slots)
}

func (h *AppointmentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "appointment-service", "status": "healthy"})
}

func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	userID, _ := middleware.CallerID(c)
	role, ok := middleware.CallerRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
		return
	}

	var req model.AppointmentStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !model.IsValidAppointmentStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment status; allowed values: pending, confirmed, cancelled, completed"})
		return
	}

	if err := h.svc.UpdateAppointmentStatus(id, userID, role, req.Status); err != nil {
		if strings.Contains(err.Error(), "only doctors or admins can update appointment status") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "forbidden: not your appointment") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *AppointmentHandler) CreateSlot(c *gin.Context) {
	userID, ok := middleware.CallerID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	role, ok := middleware.CallerRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
		return
	}

	var req model.CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, ok := middleware.CallerToken(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller token"})
		return
	}

	slot, err := h.svc.CreateSlot(userID, token, role, &req)
	if err != nil {
		if strings.Contains(err.Error(), "end time must be after start time") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "hospital is required") || strings.Contains(err.Error(), "doctor_id is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot be in the past") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "slot overlaps an existing slot for this doctor") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "forbidden: you can only create slots for your own doctor profile") ||
			strings.Contains(err.Error(), "doctor profile not found for current user") ||
			strings.Contains(err.Error(), "doctor service unavailable") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "add your hospital") ||
			strings.Contains(err.Error(), "hospital must match") ||
			strings.Contains(err.Error(), "doctor_id is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, slot)
}

func (h *AppointmentHandler) UpdateSlot(c *gin.Context) {
	id := c.Param("id")
	userID, _ := middleware.CallerID(c)
	role, ok := middleware.CallerRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
		return
	}

	var req model.UpdateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := h.svc.UpdateSlot(id, userID, role, &req)
	if err != nil {
		if strings.Contains(err.Error(), "slot not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot update a booked slot") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "end time must be after start time") ||
			strings.Contains(err.Error(), "add your hospital") ||
			strings.Contains(err.Error(), "hospital must match") ||
			strings.Contains(err.Error(), "cannot be in the past") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "hospital is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "slot overlaps an existing slot for this doctor") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slot)
}

func (h *AppointmentHandler) DeleteSlot(c *gin.Context) {
	id := c.Param("id")
	userID, _ := middleware.CallerID(c)
	role, ok := middleware.CallerRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller role"})
		return
	}

	if err := h.svc.DeleteSlot(id, userID, role); err != nil {
		if strings.Contains(err.Error(), "slot not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot delete a booked slot") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "slot deleted"})
}

func (h *AppointmentHandler) ListSlots(c *gin.Context) {
	userID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)

	slots, err := h.svc.ListAllSlots(userID, role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, slots)
}
