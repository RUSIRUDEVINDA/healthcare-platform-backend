package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/doctor-service/internal/middleware"
	"healthcare-platform/services/doctor-service/internal/model"
	"healthcare-platform/services/doctor-service/internal/service"
	"healthcare-platform/pkg/logger"
)

// DoctorHandler handles HTTP for doctor resources (handler layer only).
type DoctorHandler struct {
	svc *service.DoctorService
	log *logger.Logger
}

func NewDoctorHandler(svc *service.DoctorService, log *logger.Logger) *DoctorHandler {
	return &DoctorHandler{svc: svc, log: log}
}

// RegisterRoutes wires public reads and JWT-protected writes (via auth-service validate).
func (h *DoctorHandler) RegisterRoutes(router *gin.Engine, authClient *http.Client, authBaseURL string) {
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// Public: list / detail (supports ?specialization= filter per team guide)
	router.GET("/doctors", h.List)
	router.GET("/doctors/:id", h.GetByID)

	protected := router.Group("/doctors")
	protected.Use(middleware.RequireAuthViaAuthService(authClient, authBaseURL))
	{
		protected.POST("", middleware.RequireRole("admin"), h.Create)

		// Register path-param routes before PUT "" so /doctors/:id never shadows the root handler.
		protected.PUT("/:id/profile", middleware.RequireRole("doctor", "admin"), h.UpdateProfile)
		protected.PUT("/:id", middleware.RequireRole("doctor", "admin"), h.Update)
		protected.PUT("", middleware.RequireRole("doctor", "admin"), h.UpdatePutRoot)

		protected.DELETE("/:id", middleware.RequireRole("admin"), h.Delete)
	}
}

func (h *DoctorHandler) List(c *gin.Context) {
	filter := c.Query("specialization")
	doctors, err := h.svc.List(filter)
	if err != nil {
		h.log.Error("List doctors failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to list doctors"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(doctors))
}

func (h *DoctorHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid doctor id"))
		return
	}

	doc, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse(err.Error()))
			return
		}
		h.log.Error("Get doctor failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to fetch doctor"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(doc))
}

func (h *DoctorHandler) Create(c *gin.Context) {
	var req model.CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	doc, err := h.svc.Create(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidNICFormat) || errors.Is(err, service.ErrInvalidSLMCFormat) {
			c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
			return
		}
		if errors.Is(err, service.ErrDuplicateNICOrSLMC) {
			c.JSON(http.StatusConflict, model.ErrorResponse(err.Error()))
			return
		}
		h.log.Error("Create doctor failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to create doctor"))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse(doc))
}

func (h *DoctorHandler) Update(c *gin.Context) {
	h.updateDoctor(c)
}

func (h *DoctorHandler) UpdateProfile(c *gin.Context) {
	h.updateDoctor(c)
}

// UpdatePutRoot handles PUT /doctors with id and fields in the JSON body.
func (h *DoctorHandler) UpdatePutRoot(c *gin.Context) {
	var req model.UpdateDoctorPutRootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}
	h.finishUpdate(c, req.ID, &req.UpdateDoctorRequest)
}

func (h *DoctorHandler) updateDoctor(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid doctor id"))
		return
	}

	var req model.UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(formatValidationError(err)))
		return
	}

	h.finishUpdate(c, id, &req)
}

func (h *DoctorHandler) finishUpdate(c *gin.Context, id int64, req *model.UpdateDoctorRequest) {
	doc, err := h.svc.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidNICFormat) || errors.Is(err, service.ErrInvalidSLMCFormat) {
			c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
			return
		}
		if errors.Is(err, service.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse(err.Error()))
			return
		}
		if errors.Is(err, service.ErrNoFieldsToUpdate) {
			c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
			return
		}
		var idInUse *service.IdentityInUseError
		if errors.As(err, &idInUse) {
			c.JSON(http.StatusConflict, model.ErrorResponse(idInUse.Error()))
			return
		}
		if errors.Is(err, service.ErrDuplicateNICOrSLMC) {
			c.JSON(http.StatusConflict, model.ErrorResponse(err.Error()))
			return
		}
		h.log.Error("Update doctor failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to update doctor"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(doc))
}

func (h *DoctorHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid doctor id"))
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, service.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse(err.Error()))
			return
		}
		h.log.Error("Delete doctor failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to delete doctor"))
		return
	}

	c.JSON(http.StatusOK, model.MessageResponse("Doctor deleted"))
}

func (h *DoctorHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "doctor-service",
	})
}

func (h *DoctorHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "doctor-service",
	})
}

func parseIDParam(c *gin.Context) (int64, error) {
	// Tolerate common mistakes: /doctors/:4, /doctors/{1} (Postman/docs placeholders sent literally).
	s := strings.TrimSpace(c.Param("id"))
	s = strings.TrimPrefix(s, ":")
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	return strconv.ParseInt(s, 10, 64)
}

func formatValidationError(err error) string {
	return "Invalid request: " + err.Error()
}
