package messaging

import (
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/payment-service/internal/service"
)

type RabbitMQEventPublisher struct {
	mqClient *rabbitmq.Client
}

func NewRabbitMQEventPublisher(mqClient *rabbitmq.Client) *RabbitMQEventPublisher {
	return &RabbitMQEventPublisher{mqClient: mqClient}
}

func (p *RabbitMQEventPublisher) PublishPaymentCompleted(event service.PaymentCompletedEvent) error {
	msg := rabbitmq.PaymentCompletedEvent{
		PaymentID:     event.PaymentID,
		AppointmentID: event.AppointmentID,
		ProviderID:    event.ProviderID,
		Timestamp:     event.Timestamp,
	}

	return p.mqClient.PublishPaymentCompleted(msg)
}

func (p *RabbitMQEventPublisher) PublishPaymentRefunded(event service.PaymentRefundedEvent) error {
	msg := rabbitmq.PaymentRefundedEvent{
		PaymentID:     event.PaymentID,
		AppointmentID: event.AppointmentID,
		Amount:        event.Amount,
		Timestamp:     event.Timestamp,
	}

	return p.mqClient.PublishPaymentRefunded(msg)
}
