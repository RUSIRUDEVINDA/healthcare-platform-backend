package model

import "time"

type Doctor struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Specialization string    `json:"specialization"`
	Experience     int       `json:"experience"`
	Hospital       string    `json:"hospital"`
	NIC            string    `json:"nic"`
	SLMCNo         string    `json:"slmc_no"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateDoctorRequest struct {
	Name           string `json:"name" binding:"required,min=2,max=255"`
	Specialization string `json:"specialization" binding:"required,min=2,max=255"`
	Experience     int    `json:"experience" binding:"required,min=0,max=80"`
	Hospital       string `json:"hospital" binding:"required,min=2,max=255"`
	NIC            string `json:"nic" binding:"required,len=12"`
	SLMCNo         string `json:"slmc_no" binding:"required,len=5"`
}

type UpdateDoctorRequest struct {
	Name           *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Specialization *string `json:"specialization,omitempty" binding:"omitempty,min=2,max=255"`
	Experience     *int    `json:"experience,omitempty" binding:"omitempty,min=0,max=80"`
	Hospital       *string `json:"hospital,omitempty" binding:"omitempty,min=2,max=255"`
	NIC            *string `json:"nic,omitempty" binding:"omitempty,len=12"`
	SLMCNo         *string `json:"slmc_no,omitempty" binding:"omitempty,len=5"`
}

// UpdateDoctorPutRootRequest is the body for PUT /doctors (doctor id in JSON).
type UpdateDoctorPutRootRequest struct {
	ID int64 `json:"id" binding:"required"`
	UpdateDoctorRequest
}

// ValidateTokenResponse matches auth-service GET /auth/validate JSON body.
type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

// APIResponse is the standard JSON envelope for doctor-service handlers.
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
