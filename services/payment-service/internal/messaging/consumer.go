package messaging

import (
	"encoding/json"
	"fmt"
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
	queueName := "payment_service_queue"

	// Existing: Handle Booked
	err := c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentBooked,
		rabbitmq.RoutingKeyAppointmentBooked,
	)
	if err != nil {
		return fmt.Errorf("messaging.Start: booked: %w", err)
	}

	// New: Handle Cancelled
	err = c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentCancelled,
		rabbitmq.RoutingKeyAppointmentCancelled,
	)
	if err != nil {
		return fmt.Errorf("messaging.Start: cancelled: %w", err)
	}

	c.log.Info("Payment service consumer started")
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

	c.log.Info("Processing appointment.cancelled event", "appointment_id", event.AppointmentID)

	if err := c.svc.CancelPaymentByAppointmentID(event.AppointmentID); err != nil {
		return fmt.Errorf("messaging.handleAppointmentCancelled service: %w", err)
	}

	return nil
}
