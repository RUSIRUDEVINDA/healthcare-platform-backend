package model

import "time"

type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Role       Role      `json:"role"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	IsVerified bool      `json:"is_verified"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DoctorVerification struct {
	DoctorID   string     `json:"doctor_id"`
	VerifiedBy string     `json:"verified_by,omitempty"`
	Notes      string     `json:"notes,omitempty"`
	Status     string     `json:"status"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Appointment struct {
	ID          string    `json:"id"`
	PatientID   string    `json:"patient_id"`
	DoctorID    string    `json:"doctor_id"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Reason      string    `json:"reason,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	Provider  string    `json:"provider"`
	Reference string    `json:"reference,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VerifyDoctorRequest struct {
	Notes string `json:"notes"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func ErrorResponse(err string) APIResponse {
	return APIResponse{Success: false, Error: err}
}

func MessageResponse(msg string) APIResponse {
	return APIResponse{Success: true, Message: msg}
}
