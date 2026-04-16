package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/payment-service/internal/model"
	"healthcare-platform/services/payment-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	svc PaymentService
	log *logger.Logger
}

type PaymentService interface {
	CreatePayment(req *model.CreatePaymentRequest) (*model.PaymentResponse, error)
	Checkout(req *model.CheckoutRequest) (*model.CheckoutResponse, error)
	HandlePayHereNotification(n *model.PayHereNotification) error
	GetPaymentByID(id string) (*model.Payment, error)
	ListPaymentsByPatient(patientID string) ([]*model.Payment, error)
}

func NewPaymentHandler(svc PaymentService, log *logger.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, log: log}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.Engine) {
	payments := router.Group("/api/v1/payments")
	{
		payments.POST("/", h.CreatePayment)
		payments.POST("/checkout", h.CheckoutPayment)
		payments.POST("/webhook/payhere", h.PayHereWebhook)
		payments.GET("/:id", h.GetPayment)
		payments.GET("/patient/:patient_id", h.ListPayments)
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

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	patientID := c.Param("patient_id")
	if _, err := uuid.Parse(patientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID format"})
		return
	}

	payments, err := h.svc.ListPaymentsByPatient(patientID)
	if err != nil {
		h.log.Error("Failed to list payments", "patient_id", patientID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list payments"})
		return
	}

	if payments == nil {
		payments = []*model.Payment{}
	}

	c.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) CheckoutPayment(c *gin.Context) {
	var req model.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Checkout(&req)
	if err != nil {
		h.log.Error("Failed to create checkout", "error", err)
		status := http.StatusInternalServerError
		msg := "Failed to create checkout"
		errMsg := err.Error()
		if strings.Contains(errMsg, "required") || strings.Contains(errMsg, "not configured") || strings.Contains(errMsg, "not found") {
			status = http.StatusBadRequest
			msg = "Invalid checkout request"
		}
		if strings.Contains(errMsg, "already completed") {
			status = http.StatusConflict
			msg = "Payment already completed"
		}

		if gin.Mode() != gin.ReleaseMode {
			c.JSON(status, gin.H{"error": msg, "details": errMsg})
			return
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) PayHereWebhook(c *gin.Context) {
	var n model.PayHereNotification
	if err := c.ShouldBind(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.HandlePayHereNotification(&n); err != nil {
		h.log.Error("PayHere webhook processing failed", "order_id", n.OrderID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "payment-service", "status": "healthy"})
}
