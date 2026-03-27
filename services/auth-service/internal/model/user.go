package model

import "time"

// Role defines user roles in the system
type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
	RoleAdmin   Role = "admin"
)

// User is the core domain model stored in the database
type User struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         Role      `db:"role"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	IsVerified   bool      `db:"is_verified"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ──────────────────────────────────────────────
// Request DTOs (Data Transfer Objects)
// These come FROM the client (JSON body)
// ──────────────────────────────────────────────

type RegisterRequest struct {
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8,max=72"`
	FirstName string `json:"first_name" binding:"required,min=2,max=50"`
	LastName  string `json:"last_name"  binding:"required,min=2,max=50"`
	Role      Role   `json:"role"       binding:"required,oneof=patient doctor"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ──────────────────────────────────────────────
// Response DTOs
// These go TO the client
// ──────────────────────────────────────────────

type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"` // seconds
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Role       Role   `json:"role"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	IsVerified bool   `json:"is_verified"`
}

// ValidateTokenResponse is returned to API Gateway / other services
// when they call GET /auth/validate to check a JWT
type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   Role   `json:"role,omitempty"`
}

// APIResponse is a generic wrapper for all HTTP responses
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
