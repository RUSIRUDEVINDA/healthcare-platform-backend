package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

// loadDotEnv loads .env from the process cwd, then from this service's directory
// (services/api-gateway/.env) so `go run` from the repo root still picks up gateway env.
func loadDotEnv() {
	_ = godotenv.Load()
	if _, file, _, ok := runtime.Caller(0); ok {
		svcRoot := filepath.Join(filepath.Dir(file), "..", "..")
		_ = godotenv.Load(filepath.Join(svcRoot, ".env"))
	}
}

// Config holds upstream base URLs for the API gateway (team-guide: port 8000, /api/* routing).
type Config struct {
	Port   string
	AppEnv string

	AuthServiceURL          string
	DoctorServiceURL        string
	AISymptomServiceURL     string
	PatientServiceURL       string
	AppointmentServiceURL   string
	PaymentServiceURL       string
	AdminServiceURL         string
	TelemedicineServiceURL  string
	FileStorageServiceURL   string
	NotificationServiceURL  string
	SupportServiceURL       string
}

// Load reads configuration from environment (.env optional).
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		Port:                   getEnv("PORT", "8000"),
		AppEnv:                 getEnv("APP_ENV", "development"),
		AuthServiceURL:         strings.TrimRight(getEnv("AUTH_SERVICE_URL", "http://auth-service:8001"), "/"),
		DoctorServiceURL:       strings.TrimRight(getEnv("DOCTOR_SERVICE_URL", "http://doctor-service:8003"), "/"),
		AISymptomServiceURL:    strings.TrimRight(getEnv("AI_SYMPTOM_SERVICE_URL", "http://ai-symptom-service:8008"), "/"),
		PatientServiceURL:      strings.TrimRight(getEnv("PATIENT_SERVICE_URL", "http://patient-service:8002"), "/"),
		AppointmentServiceURL:  strings.TrimRight(getEnv("APPOINTMENT_SERVICE_URL", "http://appointment-service:8004"), "/"),
		PaymentServiceURL:      strings.TrimRight(getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8005"), "/"),
		AdminServiceURL:        strings.TrimRight(getEnv("ADMIN_SERVICE_URL", "http://admin-service:8007"), "/"),
		TelemedicineServiceURL: strings.TrimRight(getEnv("TELEMEDICINE_SERVICE_URL", "http://telemedicine-service:8009"), "/"),
		FileStorageServiceURL:  strings.TrimRight(getEnv("FILE_STORAGE_SERVICE_URL", "http://file-storage-service:8010"), "/"),
		NotificationServiceURL: strings.TrimRight(getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8006"), "/"),
		SupportServiceURL:      strings.TrimRight(getEnv("SUPPORT_SERVICE_URL", "http://support-service:8011"), "/"),
	}

	for _, u := range []struct {
		name string
		val  string
	}{
		{"AUTH_SERVICE_URL", cfg.AuthServiceURL},
		{"DOCTOR_SERVICE_URL", cfg.DoctorServiceURL},
		{"AI_SYMPTOM_SERVICE_URL", cfg.AISymptomServiceURL},
		{"PATIENT_SERVICE_URL", cfg.PatientServiceURL},
		{"APPOINTMENT_SERVICE_URL", cfg.AppointmentServiceURL},
		{"PAYMENT_SERVICE_URL", cfg.PaymentServiceURL},
		{"ADMIN_SERVICE_URL", cfg.AdminServiceURL},
		{"TELEMEDICINE_SERVICE_URL", cfg.TelemedicineServiceURL},
		{"FILE_STORAGE_SERVICE_URL", cfg.FileStorageServiceURL},
		{"NOTIFICATION_SERVICE_URL", cfg.NotificationServiceURL},
		{"SUPPORT_SERVICE_URL", cfg.SupportServiceURL},
	} {
		if _, err := url.Parse(u.val); err != nil {
			return nil, fmt.Errorf("%s: %w", u.name, err)
		}
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
