package model

import (
	"time"
)

type Patient struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Email            string     `json:"email"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	Gender           *string    `json:"gender,omitempty"`
	PhoneNumber      *string    `json:"phone_number,omitempty"`
	Address          *string    `json:"address,omitempty"`
	EmergencyContact *string    `json:"emergency_contact,omitempty"`
	BloodGroup       *string    `json:"blood_group,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type UpdatePatientRequest struct {
	DateOfBirth      *FlexibleTime `json:"date_of_birth"`
	Gender           *string       `json:"gender"`
	PhoneNumber      *string       `json:"phone_number"`
	Address          *string       `json:"address"`
	EmergencyContact *string       `json:"emergency_contact"`
	BloodGroup       *string       `json:"blood_group"`
}

type PatchPatientRequest struct {
	DateOfBirth      *FlexibleTime `json:"date_of_birth,omitempty"`
	Gender           *string       `json:"gender,omitempty"`
	PhoneNumber      *string       `json:"phone_number,omitempty"`
	Address          *string       `json:"address,omitempty"`
	EmergencyContact *string       `json:"emergency_contact,omitempty"`
	BloodGroup       *string       `json:"blood_group,omitempty"`
}
