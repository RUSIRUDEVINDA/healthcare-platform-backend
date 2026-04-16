package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/appointment-service/internal/service"
)

type AppointmentConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.AppointmentService
	log      *logger.Logger
}

func NewAppointmentConsumer(mqClient *rabbitmq.Client, svc *service.AppointmentService, log *logger.Logger) *AppointmentConsumer {
	return &AppointmentConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *AppointmentConsumer) Start() error {
	// Queue for payment completion events
	paymentQueueName := "appointment_payment_completed_queue"
	if err := c.mqClient.ConsumeQueue(
		paymentQueueName,
		rabbitmq.ExchangePaymentEvents,
		c.handlePaymentCompleted,
		rabbitmq.RoutingKeyPaymentCompleted,
	); err != nil {
		return fmt.Errorf("messaging.Start payment: %w", err)
	}

	// Queue for patient deletion events
	patientDeletedQueueName := "appointment_patient_deleted_queue"
	if err := c.mqClient.ConsumeQueue(
		patientDeletedQueueName,
		rabbitmq.ExchangeUserEvents,
		c.handlePatientDeleted,
		rabbitmq.RoutingKeyPatientDeleted,
	); err != nil {
		return fmt.Errorf("messaging.Start patient.deleted: %w", err)
	}

	c.log.Info("Appointment service consumer started successfully")
	return nil
}

func (c *AppointmentConsumer) handlePaymentCompleted(body []byte) error {
	var event rabbitmq.PaymentCompletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePaymentCompleted unmarshal: %w", err)
	}

	c.log.Info("Processing payment.completed event", "appointment_id", event.AppointmentID)

	return c.svc.HandlePaymentCompleted(event.AppointmentID)
}

func (c *AppointmentConsumer) handlePatientDeleted(body []byte) error {
	var event rabbitmq.PatientDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePatientDeleted unmarshal: %w", err)
	}

	c.log.Info("Processing patient.deleted event", "patient_id", event.PatientID)

	return c.svc.DeletePatientAppointments(event.PatientID)
}
