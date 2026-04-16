package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	AppEnv                string
	DatabaseURL           string
	JWTSecret             string
	AppointmentServiceURL string
	JitsiBaseURL          string
	RabbitMQURL           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8009"),
		AppEnv:                getEnv("APP_ENV", "development"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		AppointmentServiceURL: getEnv("APPOINTMENT_SERVICE_URL", "http://localhost:8004"),
		JitsiBaseURL:          getEnv("JITSI_BASE_URL", "https://meet.jit.si"),
		RabbitMQURL:           getEnv("RABBITMQ_URL", "amqp://admin:password123@localhost:5672/"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.AppointmentServiceURL == "" {
		return errors.New("APPOINTMENT_SERVICE_URL is required")
	}
	if c.JitsiBaseURL == "" {
		return errors.New("JITSI_BASE_URL is required")
	}
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
