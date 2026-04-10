package model

import (
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID            string        `json:"id"`
	AppointmentID string        `json:"appointment_id"`
	PatientID     string        `json:"patient_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	Provider      string        `json:"provider"` // e.g., "stripe", "paypal"
	ProviderID    string        `json:"provider_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreatePaymentRequest struct {
	AppointmentID string  `json:"appointment_id"`
	PatientID     string  `json:"patient_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

type PaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}
