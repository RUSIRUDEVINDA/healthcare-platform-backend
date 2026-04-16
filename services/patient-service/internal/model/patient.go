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
	Nationality      *string    `json:"nationality,omitempty"`
	NIC              *string    `json:"nic,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type UpdatePatientRequest struct {
	DateOfBirth      *FlexibleTime `json:"date_of_birth" binding:"omitempty"`
	Gender           *string       `json:"gender"        binding:"omitempty,oneof=male female other"`
	PhoneNumber      *string       `json:"phone_number"  binding:"omitempty,min=9,max=20"`
	Address          *string       `json:"address"       binding:"omitempty,min=5,max=500"`
	EmergencyContact *string       `json:"emergency_contact" binding:"omitempty,min=9,max=20"`
	BloodGroup       *string       `json:"blood_group"   binding:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O-"`
	Nationality      *string       `json:"nationality"   binding:"omitempty,min=2,max=100"`
	NIC              *string       `json:"nic"           binding:"omitempty,min=10,max=12"`
}

type PatchPatientRequest struct {
	DateOfBirth      *FlexibleTime `json:"date_of_birth,omitempty" binding:"omitempty"`
	Gender           *string       `json:"gender,omitempty"        binding:"omitempty,oneof=male female other"`
	PhoneNumber      *string       `json:"phone_number,omitempty"  binding:"omitempty,min=9,max=20"`
	Address          *string       `json:"address,omitempty"       binding:"omitempty,min=5,max=500"`
	EmergencyContact *string       `json:"emergency_contact,omitempty" binding:"omitempty,min=9,max=20"`
	BloodGroup       *string       `json:"blood_group,omitempty"   binding:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O-"`
	Nationality      *string       `json:"nationality,omitempty"   binding:"omitempty,min=2,max=100"`
	NIC              *string       `json:"nic,omitempty"           binding:"omitempty,min=10,max=12"`
}
