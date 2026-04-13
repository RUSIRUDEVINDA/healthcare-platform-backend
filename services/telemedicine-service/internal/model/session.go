package model

import "time"

type SessionStatus string

type ConsultationMode string

const (
	SessionScheduled SessionStatus = "scheduled"
	SessionActive    SessionStatus = "active"
	SessionEnded     SessionStatus = "ended"
	SessionCancelled SessionStatus = "cancelled"
)

const (
	ConsultationModePhysical ConsultationMode = "physical"
	ConsultationModeJitsi    ConsultationMode = "jitsi"
)

var allowedSessionStatuses = map[SessionStatus]struct{}{
	SessionScheduled: {},
	SessionActive:    {},
	SessionEnded:     {},
	SessionCancelled: {},
}

func IsValidSessionStatus(status SessionStatus) bool {
	_, ok := allowedSessionStatuses[status]
	return ok
}

type Session struct {
	ID                string        `json:"id"`
	AppointmentID     string        `json:"appointment_id"`
	SessionNumber     int           `json:"session_number"`
	PatientID         string        `json:"patient_id"`
	DoctorID          string        `json:"doctor_id"`
	DoctorOwnerUserID string        `json:"doctor_owner_user_id"`
	RoomName          string        `json:"room_name"`
	JoinURL           string        `json:"join_url"`
	Provider          string        `json:"provider"`
	Status            SessionStatus `json:"status"`
	Purpose           string        `json:"purpose,omitempty"`
	Notes             string        `json:"notes,omitempty"`
	CreatedByUserID   string        `json:"created_by_user_id"`
	CreatedByRole     string        `json:"created_by_role"`
	StartedAt         *time.Time    `json:"started_at,omitempty"`
	EndedAt           *time.Time    `json:"ended_at,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type CreateSessionRequest struct {
	Purpose string `json:"purpose"`
	Notes   string `json:"notes"`
}

type SessionJoinResponse struct {
	SessionID     string        `json:"session_id"`
	AppointmentID string        `json:"appointment_id"`
	RoomName      string        `json:"room_name"`
	JoinURL       string        `json:"join_url"`
	Provider      string        `json:"provider"`
	Status        SessionStatus `json:"status"`
	SessionNumber int           `json:"session_number"`
}
