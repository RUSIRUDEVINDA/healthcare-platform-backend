package handler

import "testing"

func TestMatchesPrefix(t *testing.T) {
	tests := []struct {
		path, prefix string
		want         bool
	}{
		{"/api/auth/login", "/api/auth", true},
		{"/api/auth", "/api/auth", true},
		{"/api/auth?x=1", "/api/auth", true},
		{"/api/authentication", "/api/auth", false},
		{"/api/ai/symptom/check", "/api/ai/symptom", true},
		{"/api/ai/symptom", "/api/ai/symptom", true},
		{"/api/ai/foo", "/api/ai/symptom", false},
		{"/api/doctors/5", "/api/doctors", true},
		{"/api/v1/doctors/5", "/api/v1/doctors", true},
		{"/api/v1/doctors", "/api/v1/doctors", true},
		{"/api/v1/doctorsx", "/api/v1/doctors", false},
		{"/api/v1/patient/profile", "/api/v1/patient", true},
		{"/api/v1/patient", "/api/v1/patient", true},
		{"/api/v1/patientx", "/api/v1/patient", false},
		{"/api/v1/files/patients/x/files", "/api/v1/files", true},
		{"/api/v1/appointments", "/api/v1/appointments", true},
		{"/api/v1/slots", "/api/v1/slots", true},
	}
	for _, tt := range tests {
		if got := matchesPrefix(tt.path, tt.prefix); got != tt.want {
			t.Errorf("matchesPrefix(%q,%q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
		}
	}
}
