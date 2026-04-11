package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/payment-service/internal/model"
	"healthcare-platform/services/payment-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	svc *service.PaymentService
	log *logger.Logger
}

func NewPaymentHandler(svc *service.PaymentService, log *logger.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, log: log}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.Engine) {
	payments := router.Group("/api/v1/payments")
	{
		payments.POST("/", h.CreatePayment)
		payments.GET("/:id", h.GetPayment)
	}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreatePayment(&req)
	if err != nil {
		var alreadyExistsErr *service.PaymentAlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			c.JSON(http.StatusConflict, gin.H{
				"error":          "Payment already exists for this appointment",
				"appointment_id": alreadyExistsErr.AppointmentID,
				"payment_id":     alreadyExistsErr.ExistingID,
			})
			return
		}

		h.log.Error("Failed to create payment", "error", err)
		if gin.Mode() != gin.ReleaseMode {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment", "details": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID format"})
		return
	}

	p, err := h.svc.GetPaymentByID(id)
	if err != nil {
		h.log.Error("Failed to fetch payment", "payment_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment"})
		return
	}

	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "payment-service", "status": "healthy"})
}
