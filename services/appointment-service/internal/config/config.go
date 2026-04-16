package config

import (
	"os"
)

type Config struct {
	AppEnv             string
	Port               string
	DatabaseURL        string
	RabbitMQURL        string
	JWTSecret          string
	DoctorServiceURL   string
	JitsiBaseURL       string
	PatientServiceURL  string
	InternalAPIKey     string // shared with patient-service for internal name lookup
}

func Load() (*Config, error) {
	return &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8004"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://user:password@localhost:5435/appointment_db?sslmode=disable"),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://admin:password123@localhost:5672/"),
		JWTSecret:         getEnv("JWT_SECRET", "local-dev-secret-change-this-in-production-32chars"),
		DoctorServiceURL:  getEnv("DOCTOR_SERVICE_URL", "http://doctor-service:8003"),
		JitsiBaseURL:      getEnv("JITSI_BASE_URL", "https://meet.jit.si"),
		PatientServiceURL: getEnv("PATIENT_SERVICE_URL", "http://patient-service:8002"),
		InternalAPIKey:    getEnv("INTERNAL_API_KEY", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
