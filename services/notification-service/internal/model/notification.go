package model

type DeliveryChannel string

const (
	DeliveryChannelEmail DeliveryChannel = "email"
	DeliveryChannelSMS   DeliveryChannel = "sms"
)

type NotificationPayload struct {
	Recipient string          `json:"recipient"`
	Subject   string          `json:"subject,omitempty"`
	Body      string          `json:"body"`
	Channel   DeliveryChannel `json:"channel"`
}
