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
	// Root-level queue name for the payment completion handler.
	// This keeps appointment payment updates isolated from other consumers.
	queueName := "appointment_payment_completed_queue"

	if err := c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangePaymentEvents,
		c.handlePaymentCompleted,
		rabbitmq.RoutingKeyPaymentCompleted,
	); err != nil {
		return fmt.Errorf("messaging.Start: %w", err)
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
