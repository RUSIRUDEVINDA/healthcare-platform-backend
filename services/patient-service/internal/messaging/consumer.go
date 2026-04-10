package messaging

import (
	"encoding/json"
	"fmt"

	"healthcare-platform/services/patient-service/internal/service"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/pkg/logger"
)

type PatientConsumer struct {
	mqClient *rabbitmq.Client
	svc      *service.PatientService
	log      *logger.Logger
}

func NewPatientConsumer(mqClient *rabbitmq.Client, svc *service.PatientService, log *logger.Logger) *PatientConsumer {
	return &PatientConsumer{mqClient: mqClient, svc: svc, log: log}
}

func (c *PatientConsumer) Start() error {
	// Root-level queue name for the patient profile creation
	// Since multiple services might listen to "user.registered", we give each its own queue name
	queueName := "patient_profile_creator_queue"

	err := c.mqClient.ConsumeQueue(
		queueName,
		rabbitmq.ExchangeUserEvents,
		c.handleUserRegistered,
		rabbitmq.RoutingKeyUserRegistered,
	)

	if err != nil {
		return fmt.Errorf("messaging.Start: %w", err)
	}

	c.log.Info("Patient service consumer started successfully")
	return nil
}

func (c *PatientConsumer) handleUserRegistered(body []byte) error {
	var event rabbitmq.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("consumer.handleUserRegistered unmarshal: %w", err)
	}

	// Only create patient profiles for users with the 'patient' role
	if event.Role != "patient" {
		c.log.Info("Ignoring user.registered event (not a patient)", "user_id", event.UserID, "role", event.Role)
		return nil
	}

	c.log.Info("Processing user.registered event", "user_id", event.UserID)

	// Business logic: create the profile
	return c.svc.CreateFromUserEvent(event.UserID, event.Email, event.FirstName, event.LastName)
}
