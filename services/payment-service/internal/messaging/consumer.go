package messaging

import (
	"encoding/json"
	"fmt"
	"strings"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/payment-service/internal/model"
	"healthcare-platform/services/payment-service/internal/service"
)

type PaymentConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.PaymentService
	log      *logger.Logger
}

func NewPaymentConsumer(mqClient *rabbitmq.Client, svc *service.PaymentService, log *logger.Logger) *PaymentConsumer {
	return &PaymentConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *PaymentConsumer) Start() error {
	// Build unique queue names per event type to prevent competing consumers
	// from "stealing" messages with different schemas.
	
	// Handle Booked
	err := c.mqClient.ConsumeQueue(
		"payment_service_booked_queue",
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentBooked,
		rabbitmq.RoutingKeyAppointmentBooked,
	)
	if err != nil {
		return fmt.Errorf("messaging.Start: booked: %w", err)
	}

	// Handle Cancelled
	err = c.mqClient.ConsumeQueue(
		"payment_service_cancelled_queue",
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentCancelled,
		rabbitmq.RoutingKeyAppointmentCancelled,
	)
	if err != nil {
		return fmt.Errorf("messaging.Start: cancelled: %w", err)
	}

	// Handle Patient Deleted
	err = c.mqClient.ConsumeQueue(
		"payment_service_patient_deleted_queue",
		rabbitmq.ExchangeUserEvents,
		c.handlePatientDeleted,
		rabbitmq.RoutingKeyPatientDeleted,
	)
	if err != nil {
		return fmt.Errorf("messaging.Start: patient.deleted: %w", err)
	}

	c.log.Info("Payment service consumers started with dedicated queues")
	return nil
}

func (c *PaymentConsumer) handleAppointmentBooked(body []byte) error {
	var event rabbitmq.AppointmentBookedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("messaging.handleAppointmentBooked unmarshal: %w", err)
	}

	c.log.Info("Processing appointment.booked event", "appointment_id", event.AppointmentID)

	// Automatically create a pending payment record
	req := &model.CreatePaymentRequest{
		AppointmentID: event.AppointmentID,
		PatientID:     event.PatientID,
		Amount:        event.ConsultFee,
		Currency:      "LKR", // Default for local consultations
	}

	_, err := c.svc.CreatePayment(req)
	if err != nil {
		return fmt.Errorf("messaging.handleAppointmentBooked create payment: %w", err)
	}

	return nil
}

func (c *PaymentConsumer) handleAppointmentCancelled(body []byte) error {
	var event rabbitmq.AppointmentCancelledEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("messaging.handleAppointmentCancelled unmarshal: %w", err)
	}

	c.log.Info("Processing appointment.cancelled event", "appointment_id", event.AppointmentID, "cancelled_by", event.CancelledBy)

	if strings.ToLower(event.CancelledBy) == "doctor" {
		if err := c.svc.RefundPayment(event.AppointmentID); err != nil {
			return fmt.Errorf("messaging.handleAppointmentCancelled refund: %w", err)
		}
	} else {
		if err := c.svc.CancelPaymentByAppointmentID(event.AppointmentID); err != nil {
			return fmt.Errorf("messaging.handleAppointmentCancelled service: %w", err)
		}
	}

	return nil
}

func (c *PaymentConsumer) handlePatientDeleted(body []byte) error {
	var event rabbitmq.PatientDeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePatientDeleted unmarshal: %w", err)
	}

	c.log.Info("Processing patient.deleted event", "patient_id", event.PatientID)
	// Keep payment history intact even if a patient account is removed.
	// Other services may still clean up their own data, but payments are retained
	// for audit/history purposes.
	c.log.Warn("Skipping hard delete of payment history for patient.deleted event", "patient_id", event.PatientID)
	return nil
}
