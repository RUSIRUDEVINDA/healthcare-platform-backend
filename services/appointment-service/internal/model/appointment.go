package model

import "time"

type AppointmentStatus string

type PaymentStatus string

type PaymentMode string

type ConsultationMode string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusCompleted AppointmentStatus = "completed"
)

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
	PaymentOverdue PaymentStatus = "overdue"
	PaymentFailed  PaymentStatus = "failed"
	PaymentExpired PaymentStatus = "expired"
)

const (
	PaymentModePayNow   PaymentMode = "pay_now"
	PaymentModePayLater PaymentMode = "pay_later"
)

const (
	ConsultationModePhysical ConsultationMode = "physical"
	ConsultationModeJitsi    ConsultationMode = "jitsi"
)

var allowedAppointmentStatuses = map[AppointmentStatus]struct{}{
	StatusPending:   {},
	StatusConfirmed: {},
	StatusCancelled: {},
	StatusCompleted: {},
}

var allowedPaymentStatuses = map[PaymentStatus]struct{}{
	PaymentPending: {},
	PaymentPaid:    {},
	PaymentOverdue: {},
	PaymentFailed:  {},
	PaymentExpired: {},
}

func IsValidAppointmentStatus(status AppointmentStatus) bool {
	_, ok := allowedAppointmentStatuses[status]
	return ok
}

func IsValidPaymentStatus(status PaymentStatus) bool {
	_, ok := allowedPaymentStatuses[status]
	return ok
}

func IsValidConsultationMode(mode ConsultationMode) bool {
	return mode == ConsultationModePhysical || mode == ConsultationModeJitsi
}

type Appointment struct {
	ID                string            `json:"id"`
	PatientID         string            `json:"patient_id"`
	DoctorID          string            `json:"doctor_id"`
	DoctorOwnerUserID string            `json:"doctor_owner_user_id"`
	SlotID            string            `json:"slot_id"`
	ConsultationMode  ConsultationMode  `json:"consultation_mode"`
	RoomName          string            `json:"room_name,omitempty"`
	JoinURL           string            `json:"join_url,omitempty"`
	ScheduledAt       time.Time         `json:"scheduled_at"`
	DurationMinutes   int               `json:"duration_minutes"`
	Status            AppointmentStatus `json:"status"`
	PaymentStatus     PaymentStatus     `json:"payment_status"`
	PaymentDueAt      *time.Time        `json:"payment_due_at,omitempty"`
	PaidAt            *time.Time        `json:"paid_at,omitempty"`
	Notes             string            `json:"notes"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type Slot struct {
	ID          string    `json:"id"`
	DoctorID    string    `json:"doctor_id"`
	OwnerUserID string    `json:"owner_user_id"`
	Hospital    string    `json:"hospital"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsBooked    bool      `json:"is_booked"`
}

// ---- Request / Response DTOs ----

type BookAppointmentRequest struct {
	SlotID           string           `json:"slot_id,omitempty"`
	DoctorID         string           `json:"doctor_id,omitempty"`
	ScheduledAt      *time.Time       `json:"scheduled_at,omitempty"`
	DurationMinutes  *int             `json:"duration_minutes,omitempty"`
	Notes            string           `json:"notes"`
	PaymentMode      PaymentMode      `json:"payment_mode,omitempty"`
	ConsultationMode ConsultationMode `json:"consultation_mode,omitempty"`
}

type UpdateAppointmentRequest struct {
	ScheduledAt     *time.Time `json:"scheduled_at"`
	DurationMinutes *int       `json:"duration_minutes"`
	Notes           *string    `json:"notes"`
}

type AppointmentStatusUpdateRequest struct {
	Status AppointmentStatus `json:"status" binding:"required"`
}

type CreateSlotRequest struct {
	DoctorID  string    `json:"doctor_id" binding:"required"`
	Hospital  string    `json:"hospital" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

type UpdateSlotRequest struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	IsBooked  *bool      `json:"is_booked"`
	Hospital  *string    `json:"hospital"`
}

// ---- RabbitMQ Event ----

type AppointmentEvent struct {
	AppointmentID string    `json:"appointment_id"`
	PatientID     string    `json:"patient_id"`
	DoctorID      string    `json:"doctor_id"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	Timestamp     time.Time `json:"timestamp"`
}
