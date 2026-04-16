package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/file-storage-service/internal/service"
)

type FileEventConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.FileService
	log      *logger.Logger
}

func NewFileEventConsumer(mqClient *rabbitmq.Client, svc *service.FileService, log *logger.Logger) *FileEventConsumer {
	return &FileEventConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *FileEventConsumer) Start() error {
	if err := c.mqClient.ConsumeQueue(
		"file_storage_user_deleted_queue",
		rabbitmq.ExchangeUserEvents,
		c.handleUserDeleted,
		rabbitmq.RoutingKeyUserDeleted,
	); err != nil {
		return fmt.Errorf("messaging.Start user.deleted: %w", err)
	}

	if err := c.mqClient.ConsumeQueue(
		"file_storage_patient_deleted_queue",
		rabbitmq.ExchangeUserEvents,
		c.handlePatientDeleted,
		rabbitmq.RoutingKeyPatientDeleted,
	); err != nil {
		return fmt.Errorf("messaging.Start patient.deleted: %w", err)
	}

	if err := c.mqClient.ConsumeQueue(
		"file_storage_doctor_profile_updated_queue",
		rabbitmq.ExchangeDoctorEvents,
		c.handleDoctorProfileUpdated,
		rabbitmq.RoutingKeyDoctorProfileUpdated,
	); err != nil {
		return fmt.Errorf("messaging.Start doctor.profile.updated: %w", err)
	}

	c.log.Info("File storage service consumer started")
	return nil
}

func (c *FileEventConsumer) handleUserDeleted(body []byte) error {
	var event rabbitmq.UserDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("handleUserDeleted unmarshal: %w", err)
	}

	c.log.Info("Processing user.deleted event", "user_id", event.UserID)
	return c.svc.DeleteFilesByOwner(context.Background(), event.UserID)
}

func (c *FileEventConsumer) handlePatientDeleted(body []byte) error {
	var event rabbitmq.PatientDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("handlePatientDeleted unmarshal: %w", err)
	}

	ownerID := event.PatientID
	if ownerID == "" {
		ownerID = event.UserID
	}
	c.log.Info("Processing patient.deleted event", "patient_id", event.PatientID, "user_id", event.UserID)
	return c.svc.DeleteDocumentsByOwner(context.Background(), ownerID)
}

func (c *FileEventConsumer) handleDoctorProfileUpdated(body []byte) error {
	var event rabbitmq.DoctorProfileUpdatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("handleDoctorProfileUpdated unmarshal: %w", err)
	}

	c.log.Info("Processing doctor.profile.updated event", "doctor_id", event.DoctorID, "profile_image_file_id", event.ProfileImageFileID)
	if event.ProfileImageFileID != "" {
		if _, err := c.svc.GetFile(context.Background(), event.UserID, "doctor", "", event.ProfileImageFileID); err != nil {
			c.log.Warn("Profile image reference could not be verified", "doctor_id", event.DoctorID, "error", err)
		}
	}
	return nil
}
