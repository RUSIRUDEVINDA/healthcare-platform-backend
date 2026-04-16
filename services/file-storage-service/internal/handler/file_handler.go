package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/file-storage-service/internal/middleware"
	"healthcare-platform/services/file-storage-service/internal/model"
	"healthcare-platform/services/file-storage-service/internal/repository"
	"healthcare-platform/services/file-storage-service/internal/service"
)

type FileHandler struct {
	svc *service.FileService
	log *logger.Logger
}

func NewFileHandler(svc *service.FileService, log *logger.Logger) *FileHandler {
	return &FileHandler{svc: svc, log: log}
}

func (h *FileHandler) RegisterRoutes(router *gin.Engine, auth gin.HandlerFunc) {
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api/v1")
	api.Use(auth)

	files := api.Group("/files")
	{
		files.POST("/images", middleware.RequireRoles("patient", "doctor", "admin"), h.UploadImage)
		files.PUT("/images", middleware.RequireRoles("patient", "doctor", "admin"), h.UpdateImage)
		files.POST("/documents", middleware.RequireRoles("patient", "doctor", "admin"), h.UploadDocument)
		files.POST("/patients/:patient_id/files", middleware.RequireRoles("patient", "doctor", "admin"), h.UploadPatientFile)
		files.GET("/patients/:patient_id/files", middleware.RequireRoles("patient", "doctor", "admin"), h.ListPatientFiles)
		files.GET("", middleware.RequireRoles("patient", "doctor", "admin"), h.ListMyFiles)
		files.GET("/:id", middleware.RequireRoles("patient", "doctor", "admin"), h.GetFile)
		files.GET("/:id/download", middleware.RequireRoles("patient", "doctor", "admin"), h.DownloadFile)
		files.DELETE("/:id", middleware.RequireRoles("patient", "doctor", "admin"), h.DeleteFile)
	}
}

func (h *FileHandler) UploadImage(c *gin.Context) {
	h.upload(c, true)
}

func (h *FileHandler) UpdateImage(c *gin.Context) {
	ownerID, ok := middleware.CallerID(c)
	if !ok || ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	rec, replaced, err := h.svc.UpdateImage(c.Request.Context(), ownerID, ownerID, fileHeader)
	if err != nil {
		h.writeUploadError(c, err)
		return
	}

	message := "profile image updated"
	if replaced {
		message = "profile image replaced"
	}
	c.JSON(http.StatusOK, gin.H{"file": rec, "message": message})
}

func (h *FileHandler) UploadDocument(c *gin.Context) {
	h.upload(c, false)
}

func (h *FileHandler) UploadPatientFile(c *gin.Context) {
	ownerID, ok := middleware.CallerID(c)
	if !ok || ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	role, _ := middleware.CallerRole(c)
	token, _ := middleware.CallerToken(c)
	patientID := c.Param("patient_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	docCategory := strings.TrimSpace(c.PostForm("document_category"))
	rec, replaced, err := h.svc.UploadPatientFile(c.Request.Context(), ownerID, role, token, patientID, docCategory, fileHeader)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorizedAccess) {
			h.writeAccessError(c, err)
			return
		}
		h.writeUploadError(c, err)
		return
	}
	message := "file uploaded"
	if replaced {
		message = "profile image replaced"
	}
	c.JSON(http.StatusCreated, gin.H{"file": rec, "message": message})
}

func (h *FileHandler) upload(c *gin.Context, image bool) {
	ownerID, ok := middleware.CallerID(c)
	if !ok || ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if image {
		rec, err := h.svc.UploadImage(c.Request.Context(), ownerID, ownerID, fileHeader)
		if err != nil {
			h.writeUploadError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"file": rec, "message": "profile image uploaded"})
		return
	}

	rec, err := h.svc.UploadDocument(c.Request.Context(), ownerID, ownerID, fileHeader)
	if err != nil {
		h.writeUploadError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"file": rec})
}

func (h *FileHandler) ListMyFiles(c *gin.Context) {
	ownerID, ok := middleware.CallerID(c)
	if !ok || ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}

	files, err := h.svc.ListMyFiles(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) ListPatientFiles(c *gin.Context) {
	ownerID, ok := middleware.CallerID(c)
	if !ok || ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing caller identity"})
		return
	}
	role, _ := middleware.CallerRole(c)
	token, _ := middleware.CallerToken(c)
	patientID := c.Param("patient_id")

	files, err := h.svc.ListPatientFiles(c.Request.Context(), ownerID, role, token, patientID)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) GetFile(c *gin.Context) {
	callerID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)
	token, _ := middleware.CallerToken(c)
	fileID := c.Param("id")

	file, err := h.svc.GetFile(c.Request.Context(), callerID, role, token, fileID)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"file": file})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	callerID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)
	token, _ := middleware.CallerToken(c)
	fileID := c.Param("id")

	file, err := h.svc.GetFile(c.Request.Context(), callerID, role, token, fileID)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}

	if file.StorageProvider == model.StorageProviderCloudinary {
		if file.CloudinaryURL == "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "image url unavailable"})
			return
		}
		c.Redirect(http.StatusFound, file.CloudinaryURL)
		return
	}

	reader, contentType, err := h.svc.DownloadDocument(c.Request.Context(), callerID, role, token, fileID)
	if err != nil {
		h.writeAccessError(c, err)
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+file.OriginalName+"\"")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	callerID, _ := middleware.CallerID(c)
	role, _ := middleware.CallerRole(c)
	token, _ := middleware.CallerToken(c)
	fileID := c.Param("id")

	if err := h.svc.DeleteFile(c.Request.Context(), callerID, role, token, fileID); err != nil {
		h.writeAccessError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}

func (h *FileHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "file-storage-service", "status": "healthy"})
}

func (h *FileHandler) writeUploadError(c *gin.Context, err error) {
	switch {
	case err == service.ErrProfileImageExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err == service.ErrDocumentAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err == service.ErrInvalidFileType:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case err == service.ErrFileTooLarge:
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case err == service.ErrRelationshipLookupFailed:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case strings.Contains(err.Error(), "cloudinary") || strings.Contains(err.Error(), "r2"):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
	}
}

func (h *FileHandler) writeAccessError(c *gin.Context, err error) {
	switch {
	case err == repository.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
	case err == service.ErrUnauthorizedAccess:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case err == service.ErrDocumentAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err == service.ErrRelationshipLookupFailed:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case err == service.ErrDownloadUnavailable:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
