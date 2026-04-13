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
	queueName := "telemedicine_appointment_booked_queue"

	if err := c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangeAppointmentEvents,
		c.handleAppointmentBooked,
		rabbitmq.RoutingKeyAppointmentBooked,
	); err != nil {
		return fmt.Errorf("messaging.Start: %w", err)
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
