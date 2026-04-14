package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/doctor-service/internal/service"
)

type DoctorConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.DoctorService
	log      *logger.Logger
}

func NewDoctorConsumer(mqClient *rabbitmq.Client, svc *service.DoctorService, log *logger.Logger) *DoctorConsumer {
	return &DoctorConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *DoctorConsumer) Start() error {
	// Unique queue for doctor service registration events
	queueName := "doctor_profile_creator_queue"

	err := c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangeUserEvents,
		c.handleUserRegistered,
		rabbitmq.RoutingKeyUserRegistered,
	)

	if err != nil {
		return fmt.Errorf("messaging.Start: %w", err)
	}

	c.log.Info("Doctor service consumer started successfully")
	return nil
}

func (c *DoctorConsumer) handleUserRegistered(body []byte) error {
	var event rabbitmq.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handleUserRegistered unmarshal: %w", err)
	}

	// Only create doctor profiles for users with the 'doctor' role
	if event.Role != "doctor" {
		c.log.Info("Ignoring user.registered event (not a doctor)", "user_id", event.UserID, "role", event.Role)
		return nil
	}

	c.log.Info("Processing user.registered event for doctor", "user_id", event.UserID)

	// Create the skeleton profile
	return c.svc.CreateFromUserEvent(event.UserID, event.Email, event.FirstName, event.LastName)
}
