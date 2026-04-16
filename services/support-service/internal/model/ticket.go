package model

import "time"

type TicketStatus string

const (
	StatusPending  TicketStatus = "pending"
	StatusReviewed TicketStatus = "reviewed"
	StatusResolved TicketStatus = "resolved"
	StatusRejected TicketStatus = "rejected"
)

type ReactivationTicket struct {
	ID        string       `json:"id"`
	Name      string       `json:"name" binding:"required"`
	Email     string       `json:"email" binding:"required,email"`
	Reason    string       `json:"reason" binding:"required"`
	Status    TicketStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func MessageResponse(msg string) APIResponse {
	return APIResponse{Success: true, Message: msg}
}

func ErrorResponse(err string) APIResponse {
	return APIResponse{Success: false, Error: err}
}
