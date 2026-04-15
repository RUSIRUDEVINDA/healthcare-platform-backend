package model

import (
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
	ID            string        `json:"id"`
	AppointmentID string        `json:"appointment_id"`
	PatientID     string        `json:"patient_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	Provider      string        `json:"provider"` // e.g., "payhere"
	ProviderID    string        `json:"provider_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreatePaymentRequest struct {
	AppointmentID string  `json:"appointment_id" binding:"required,uuid"`
	PatientID     string  `json:"patient_id" binding:"required,uuid"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required"`
}

type PaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Provider  string `json:"provider,omitempty"`
}

type CheckoutRequest struct {
	PaymentID     string          `json:"payment_id" binding:"omitempty,uuid"`
	AppointmentID string          `json:"appointment_id" binding:"omitempty,uuid"`
	Items         string          `json:"items"`
	Customer      PayHereCustomer `json:"customer"`
}

type PayHereCustomer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

type CheckoutResponse struct {
	PaymentID   string            `json:"payment_id"`
	Provider    string            `json:"provider"`
	CheckoutURL string            `json:"checkout_url"`
	Fields      map[string]string `json:"fields"`
}

// PayHereNotification is the POST payload sent to your notify_url (webhook).
// PayHere sends it as x-www-form-urlencoded.
type PayHereNotification struct {
	MerchantID      string `form:"merchant_id" json:"merchant_id"`
	OrderID         string `form:"order_id" json:"order_id"`
	PaymentID       string `form:"payment_id" json:"payment_id"`
	PayHereAmount   string `form:"payhere_amount" json:"payhere_amount"`
	PayHereCurrency string `form:"payhere_currency" json:"payhere_currency"`
	StatusCode      int    `form:"status_code" json:"status_code"`
	MD5Sig          string `form:"md5sig" json:"md5sig"`
	Custom1         string `form:"custom_1" json:"custom_1"`
	Custom2         string `form:"custom_2" json:"custom_2"`
	StatusMessage   string `form:"status_message" json:"status_message"`
}
