package model

// SymptomCheckRequest is the body for POST /symptoms/check.
type SymptomCheckRequest struct {
	Symptoms        string `json:"symptoms" binding:"required,min=10,max=8000"`
	OptionalContext string `json:"optional_context,omitempty" binding:"omitempty,max=2000"`
}

// SymptomCheckResponse is returned after LLM analysis (not a clinical diagnosis).
type SymptomCheckResponse struct {
	SuggestedSpecialty string `json:"suggested_specialty"`
	PreliminaryNotes   string `json:"preliminary_notes"`
	Disclaimer         string `json:"disclaimer"`
}

// ValidateTokenResponse matches auth-service GET /auth/validate JSON body.
type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

// APIResponse is the standard JSON envelope.
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
