package rabbitmq

import (
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"healthcare-platform/pkg/logger"
)

// Exchange and routing key constants
// Every service should use these same constants to avoid typos
const (
	ExchangeUserEvents        = "user_events"
	ExchangeAppointmentEvents = "appointment_events"
	ExchangePaymentEvents     = "payment_events"
	ExchangeDoctorEvents      = "doctor_events"

	RoutingKeyUserRegistered       = "user.registered"
	RoutingKeyUserLoggedIn         = "user.logged_in"
	RoutingKeyAppointmentBooked    = "appointment.booked"
	RoutingKeyAppointmentCancelled = "appointment.cancelled"
	RoutingKeyPaymentCompleted     = "payment.completed"
	RoutingKeyPaymentFailed        = "payment.failed"
	RoutingKeyDoctorCreated        = "doctor.created"
)

// Event payload structs
// All services share these definitions

// UserRegisteredEvent is published by auth-service on new registration
// Subscribers:
//   - patient-service: creates a patient profile
//   - notification-service: sends welcome email
type UserRegisteredEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Timestamp string `json:"timestamp"`
}

// AppointmentBookedEvent is published by appointment-service
// Subscribers:
//   - notification-service: sends confirmation SMS + email
//   - payment-service: creates payment record
type AppointmentBookedEvent struct {
	AppointmentID string  `json:"appointment_id"`
	PatientID     string  `json:"patient_id"`
	DoctorID      string  `json:"doctor_id"`
	PatientEmail  string  `json:"patient_email"`
	DoctorEmail   string  `json:"doctor_email"`
	ScheduledAt   string  `json:"scheduled_at"`
	ConsultFee    float64 `json:"consult_fee"`
	Timestamp     string  `json:"timestamp"`
}

// PaymentCompletedEvent is published by payment-service
// Subscribers:
//   - appointment-service: marks appointment as paid
//   - notification-service: sends payment receipt
type PaymentCompletedEvent struct {
	PaymentID     string `json:"payment_id"`
	AppointmentID string `json:"appointment_id"`
	ProviderID    string `json:"provider_id"`
	Timestamp     string `json:"timestamp"`
// DoctorCreatedEvent is published by doctor-service when a doctor record is created.
// Consumers (e.g. notification-service) can subscribe on exchange doctor_events.
type DoctorCreatedEvent struct {
	DoctorID       uint   `json:"doctor_id"`
	Name           string `json:"name"`
	Specialization string `json:"specialization"`
	Hospital       string `json:"hospital"`
	NIC            string `json:"nic"`
	SLMCNo         string `json:"slmc_no"`
	Timestamp      string `json:"timestamp"`
}

// Client
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *logger.Logger
}

func NewClient(url string, log *logger.Logger) (*Client, error) {
	var conn *amqp.Connection
	var err error

	// Retry loop — RabbitMQ container might not be ready when service starts
	for i := 1; i <= 5; i++ {
		conn, err = amqp.Dial(url)
		if err != nil {
			log.Warn("RabbitMQ connection attempt failed, retrying...", "attempt", i, "error", err)
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.NewClient: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.NewClient open channel: %w", err)
	}

	client := &Client{conn: conn, channel: ch, log: log}

	// Declare all exchanges upfront
	if err := client.declareExchanges(); err != nil {
		return nil, err
	}

	log.Info("Connected to RabbitMQ successfully")
	return client, nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// EnsureQueueBindings declares a durable queue and binds it to the given exchange/routing keys.
// This helps prevent message loss when publishers emit events before consumers start.
func (c *Client) EnsureQueueBindings(queueName, exchange string, routingKeys ...string) error {
	queue, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.EnsureQueueBindings declare queue: %w", err)
	}

	for _, routingKey := range routingKeys {
		if routingKey == "" {
			continue
		}
		if err := c.channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
			return fmt.Errorf("rabbitmq.EnsureQueueBindings bind queue: %w", err)
		}
	}

	c.log.Info("Ensured queue bindings", "queue", queueName, "exchange", exchange)
	return nil
}

// Publishers (auth-service publishes these)
func (c *Client) PublishUserRegistered(event UserRegisteredEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return c.publish(ExchangeUserEvents, RoutingKeyUserRegistered, event)
}

func (c *Client) PublishUserLoggedIn(event UserRegisteredEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return c.publish(ExchangeUserEvents, RoutingKeyUserLoggedIn, event)
}

func (c *Client) PublishPaymentCompleted(event PaymentCompletedEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return c.publish(ExchangePaymentEvents, RoutingKeyPaymentCompleted, event)
// PublishDoctorCreated publishes to the doctor_events topic exchange (routing: doctor.created).
func (c *Client) PublishDoctorCreated(event DoctorCreatedEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return c.publish(ExchangeDoctorEvents, RoutingKeyDoctorCreated, event)
}

// Consumers (used by other services)

// ConsumeQueue starts consuming messages from a queue
// handler is called for each message received
func (c *Client) ConsumeQueue(queueName, exchange string, handler func([]byte) error, routingKeys ...string) error {
	// Declare the queue
	queue, err := c.channel.QueueDeclare(
		queueName, // queue name
		true,      // durable: survives RabbitMQ restart
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.ConsumeQueue declare queue: %w", err)
	}

	// Bind queue to exchange with routing keys
	for _, routingKey := range routingKeys {
		if routingKey == "" {
			continue
		}
		if err := c.channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
			return fmt.Errorf("rabbitmq.ConsumeQueue bind queue: %w", err)
		}
	}

	// Start consuming
	msgs, err := c.channel.Consume(
		queue.Name,
		"",    // consumer tag (auto-generated)
		false, // auto-ack: false means we manually ack after processing
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.ConsumeQueue start consume: %w", err)
	}

	// Process messages in background goroutine
	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				c.log.Error("Failed to process message", "queue", queueName, "error", err)
				msg.Nack(false, true) // Requeue the message
			} else {
				msg.Ack(false) // Acknowledge successful processing
			}
		}
	}()

	c.log.Info("Started consuming queue", "queue", queueName, "exchange", exchange)
	return nil
}

// Private helpers
func (c *Client) declareExchanges() error {
	exchanges := []string{
		ExchangeUserEvents,
		ExchangeAppointmentEvents,
		ExchangePaymentEvents,
		ExchangeDoctorEvents,
	}

	for _, exchange := range exchanges {
		err := c.channel.ExchangeDeclare(
			exchange, // name
			"topic",  // type: topic routing allows wildcard patterns like "user.*"
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			return fmt.Errorf("rabbitmq.declareExchanges %s: %w", exchange, err)
		}
	}

	return nil
}

func (c *Client) publish(exchange, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rabbitmq.publish marshal: %w", err)
	}

	err = c.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Survives RabbitMQ restart
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.publish %s: %w", routingKey, err)
	}

	return nil
}
