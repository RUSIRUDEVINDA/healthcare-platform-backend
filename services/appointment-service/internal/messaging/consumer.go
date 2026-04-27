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
	// Handle Payment Completed
	if err := c.mqClient.ConsumeQueue(
		"appointment_service_payment_completed_queue",
		rabbitmq.ExchangePaymentEvents,
		c.handlePaymentCompleted,
		rabbitmq.RoutingKeyPaymentCompleted,
	); err != nil {
		return fmt.Errorf("messaging.Start payment.completed: %w", err)
	}

	// Handle Payment Refunded
	if err := c.mqClient.ConsumeQueue(
		"appointment_service_payment_refunded_queue",
		rabbitmq.ExchangePaymentEvents,
		c.handlePaymentRefunded,
		rabbitmq.RoutingKeyPaymentRefunded,
	); err != nil {
		return fmt.Errorf("messaging.Start payment.refunded: %w", err)
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

func (c *AppointmentConsumer) handlePaymentRefunded(body []byte) error {
	var event rabbitmq.PaymentRefundedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePaymentRefunded unmarshal: %w", err)
	}

	c.log.Info("Processing payment.refunded event", "appointment_id", event.AppointmentID)

	return c.svc.HandlePaymentRefunded(event.AppointmentID)
}

func (c *AppointmentConsumer) handlePatientDeleted(body []byte) error {
	var event rabbitmq.PatientDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePatientDeleted unmarshal: %w", err)
	}

	c.log.Info("Processing patient.deleted event", "patient_id", event.PatientID)

	return c.svc.DeletePatientAppointments(event.PatientID)
}
