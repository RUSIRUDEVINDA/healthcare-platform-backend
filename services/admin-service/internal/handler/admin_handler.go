package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/jwt"
	"healthcare-platform/services/admin-service/internal/middleware"
	"healthcare-platform/services/admin-service/internal/model"
	"healthcare-platform/services/admin-service/internal/service"
)

type AdminHandler struct {
	svc      *service.AdminService
	jwt      *jwt.Helper
	loggable interface {
		Info(string, ...any)
		Error(string, ...any)
	}
}

func NewAdminHandler(svc *service.AdminService, jwtHelper *jwt.Helper, log interface {
	Info(string, ...any)
	Error(string, ...any)
}) *AdminHandler {
	return &AdminHandler{svc: svc, jwt: jwtHelper, loggable: log}
}

func (h *AdminHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	auth := middleware.RequireAuth(h.jwt)
	admin := middleware.RequireAdminRole()

	group := router.Group("/admin")
	group.Use(auth, admin)
	{
		group.GET("/users", h.ListUsers)
		group.POST("/users", h.CreateUser)
		group.PUT("/users/:id", h.UpdateUser)
		group.DELETE("/users/:id", h.DeactivateUser)
		group.POST("/users/:id/reactivate", h.ReactivateUser)

		group.PUT("/doctors/:id/verify", h.VerifyDoctor)

		group.GET("/appointments", h.ListAppointments)
		group.POST("/appointments", h.CreateAppointment)
		group.PUT("/appointments/:id", h.UpdateAppointment)
		group.DELETE("/appointments/:id", h.CancelAppointment)

		group.GET("/transactions", h.ListTransactions)
		group.POST("/sync", h.SyncAll)
	}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers()
	if err != nil {
		h.loggable.Error("Failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to list users"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(users))
}

func (h *AdminHandler) VerifyDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	if doctorID == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("doctor id is required"))
		return
	}

	var req model.VerifyDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	verifiedBy, _ := c.Get(middleware.ContextUserID)
	verification, err := h.svc.VerifyDoctor(doctorID, toString(verifiedBy), req.Notes)
	if err != nil {
		h.loggable.Error("Failed to verify doctor", "doctor_id", doctorID, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to verify doctor"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(verification))
}

func (h *AdminHandler) ListAppointments(c *gin.Context) {
	appointments, err := h.svc.ListAppointments()
	if err != nil {
		h.loggable.Error("Failed to list appointments", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to list appointments"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(appointments))
}

func (h *AdminHandler) ListTransactions(c *gin.Context) {
	transactions, err := h.svc.ListTransactions()
	if err != nil {
		h.loggable.Error("Failed to list transactions", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to list transactions"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(transactions))
}

func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.DeactivateUser(userID); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse(err.Error()))
			return
		}
		h.loggable.Error("Failed to deactivate user", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to deactivate user"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("user deactivated successfully"))
}

func (h *AdminHandler) ReactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.ReactivateUser(userID); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse(err.Error()))
			return
		}
		h.loggable.Error("Failed to reactivate user", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to reactivate user"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("user reactivated successfully"))
}

func (h *AdminHandler) SyncAll(c *gin.Context) {
	synced, err := h.svc.SyncAll()
	if err != nil {
		h.loggable.Error("Sync failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("sync failed"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{"synced": synced}))
}

func (h *AdminHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "admin-service", "status": "healthy"})
}

func (h *AdminHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "admin-service", "status": "ready"})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	user, err := h.svc.CreateUser(req)
	if err != nil {
		h.loggable.Error("Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to create user"))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse(user))
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	if err := h.svc.UpdateUser(id, req); err != nil {
		h.loggable.Error("Failed to update user", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to update user"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("user updated successfully"))
}

func (h *AdminHandler) CreateAppointment(c *gin.Context) {
	var req model.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	appt, err := h.svc.CreateAppointment(req)
	if err != nil {
		h.loggable.Error("Failed to create appointment", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to create appointment"))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse(appt))
}

func (h *AdminHandler) UpdateAppointment(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	if err := h.svc.UpdateAppointment(id, req); err != nil {
		h.loggable.Error("Failed to update appointment", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to update appointment"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("appointment updated successfully"))
}

func (h *AdminHandler) CancelAppointment(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelAppointment(id); err != nil {
		h.loggable.Error("Failed to cancel appointment", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to cancel appointment"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("appointment cancelled successfully"))
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
