package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/admin-service/internal/service"
)

type Consumer struct {
	mqClient *rabbitmq.Client
	svc      *service.AdminService
	log      *logger.Logger
}

func NewConsumer(mqClient *rabbitmq.Client, svc *service.AdminService, log *logger.Logger) *Consumer {
	return &Consumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *Consumer) Start() error {
	queues := []struct {
		name       string
		exchange   string
		routingKey string
		handler    func([]byte) error
	}{
		{name: "admin_user_registered_queue", exchange: rabbitmq.ExchangeUserEvents, routingKey: rabbitmq.RoutingKeyUserRegistered, handler: c.handleUserRegistered},
		{name: "admin_appointment_booked_queue", exchange: rabbitmq.ExchangeAppointmentEvents, routingKey: rabbitmq.RoutingKeyAppointmentBooked, handler: c.handleAppointmentBooked},
		{name: "admin_payment_completed_queue", exchange: rabbitmq.ExchangePaymentEvents, routingKey: rabbitmq.RoutingKeyPaymentCompleted, handler: c.handlePaymentCompleted},
	}

	for _, queue := range queues {
		if err := c.mqClient.ConsumeQueue(queue.name, queue.exchange, queue.handler, queue.routingKey); err != nil {
			return fmt.Errorf("messaging.Start %s: %w", queue.name, err)
		}
	}

	c.log.Info("Admin service consumers started successfully")
	return nil
}

func (c *Consumer) handleUserRegistered(body []byte) error {
	var event rabbitmq.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handleUserRegistered unmarshal: %w", err)
	}
	return c.svc.HandleUserRegistered(event)
}

func (c *Consumer) handleAppointmentBooked(body []byte) error {
	var event rabbitmq.AppointmentBookedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handleAppointmentBooked unmarshal: %w", err)
	}
	if event.AppointmentID == "" || event.PatientID == "" || event.DoctorID == "" || event.ScheduledAt == "" {
		c.log.Error("Discarding invalid appointment.booked event", "reason", "missing required fields")
		return nil
	}
	return c.svc.HandleAppointmentBooked(event)
}

func (c *Consumer) handlePaymentCompleted(body []byte) error {
	var event rabbitmq.PaymentCompletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handlePaymentCompleted unmarshal: %w", err)
	}
	return c.svc.HandlePaymentCompleted(event)
}
