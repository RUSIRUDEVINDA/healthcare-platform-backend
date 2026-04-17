package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"healthcare-platform/services/support-service/internal/model"
	"healthcare-platform/services/support-service/internal/service"
)

type TicketHandler struct {
	svc *service.TicketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func (h *TicketHandler) Create(c *gin.Context) {
	var ticket model.ReactivationTicket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid request data: "+err.Error()))
		return
	}

	if err := h.svc.CreateTicket(c.Request.Context(), &ticket); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to submit ticket"))
		return
	}

	c.JSON(http.StatusCreated, model.MessageResponse("Reactivation request submitted successfully. Our admin team will review it shortly."))
}

func (h *TicketHandler) List(c *gin.Context) {
	tickets, err := h.svc.ListTickets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to fetch tickets"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(tickets))
}

func (h *TicketHandler) Resolve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Ticket ID is required"))
		return
	}

	if err := h.svc.ResolveTicket(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to resolve ticket"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("Ticket resolved successfully"))
}
