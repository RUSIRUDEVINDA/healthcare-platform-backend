package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/telemedicine-service/internal/service"
)

type AppointmentBookedConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.SessionService
	log      *logger.Logger
}

func NewAppointmentBookedConsumer(mqClient *rabbitmq.Client, svc *service.SessionService, log *logger.Logger) *AppointmentBookedConsumer {
	return &AppointmentBookedConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *AppointmentBookedConsumer) Start() error {
	// Queue for appointment booked events
	appointmentQueueName := "telemedicine_appointment_booked_queue"
	if err := c.mqClient.ConsumeQueue(
		appointmentQueueName,
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentBooked,
		rabbitmq.RoutingKeyAppointmentBooked,
	); err != nil {
		return fmt.Errorf("messaging.Start appointment.booked: %w", err)
	}

	// Queue for patient deletion events
	patientDeletedQueueName := "telemedicine_patient_deleted_queue"
	if err := c.mqClient.ConsumeQueue(
		patientDeletedQueueName,
		rabbitmq.ExchangeUserEvents,
		c.handlePatientDeleted,
		rabbitmq.RoutingKeyPatientDeleted,
	); err != nil {
		return fmt.Errorf("messaging.Start patient.deleted: %w", err)
	}

	c.log.Info("Telemedicine appointment booked consumer started successfully")
	return nil
}

func (c *AppointmentBookedConsumer) handleAppointmentBooked(body []byte) error {
	var event rabbitmq.AppointmentBookedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handleAppointmentBooked unmarshal: %w", err)
	}

	c.log.Info("Processing appointment.booked event for telemedicine", "appointment_id", event.AppointmentID, "mode", event.ConsultationMode)
	return c.svc.CreateFromAppointmentBooked(event)
}

func (c *AppointmentBookedConsumer) handlePatientDeleted(body []byte) error {
	var event rabbitmq.PatientDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePatientDeleted unmarshal: %w", err)
	}

	c.log.Info("Processing patient.deleted event for telemedicine", "patient_id", event.PatientID)
	return c.svc.DeletePatientSessions(event.PatientID)
}
