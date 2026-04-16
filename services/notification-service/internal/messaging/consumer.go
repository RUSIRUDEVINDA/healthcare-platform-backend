package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/notification-service/internal/service"
	"healthcare-platform/services/notification-service/internal/validation"
)

type Consumer struct {
	mqClient *rabbitmq.Client
	svc      *service.NotificationService
	log      *logger.Logger
}

func NewConsumer(mqClient *rabbitmq.Client, svc *service.NotificationService, log *logger.Logger) *Consumer {
	return &Consumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *Consumer) HandleAppointmentBooked(body []byte) error {
	var event rabbitmq.AppointmentBookedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.log.Error("Failed to unmarshal appointment.booked event", "error", err, "raw_body", string(body))
		return nil
	}

	if err := validation.ValidateAppointmentBookedEventLenient(&event); err != nil {
		c.log.Error("Discarding invalid appointment.booked event", "error", err, "appointment_id", event.AppointmentID, "doctor_id", event.DoctorID, "patient_id", event.PatientID)
		return nil
	}

	return c.svc.HandleAppointmentBooked(event)
}

func (c *Consumer) HandleAppointmentCancelled(body []byte) error {
	var event rabbitmq.AppointmentCancelledEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.HandleAppointmentCancelled unmarshal: %w", err)
	}
	return c.svc.HandleAppointmentCancelled(event)
}

func (c *Consumer) HandleConsultationCompleted(body []byte) error {
	var event rabbitmq.ConsultationCompletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.HandleConsultationCompleted unmarshal: %w", err)
	}
	return c.svc.HandleConsultationCompleted(event)
}
