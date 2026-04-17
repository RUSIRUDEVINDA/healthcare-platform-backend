package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	Port            string
	DatabaseURL     string
	AuthDatabaseURL        string
	AppointmentDatabaseURL string
	PaymentDatabaseURL     string
	RabbitMQURL            string
	JWTSecret              string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8007"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		AuthDatabaseURL:        getEnv("AUTH_DATABASE_URL", ""),
		AppointmentDatabaseURL: getEnv("APPOINTMENT_DATABASE_URL", ""),
		PaymentDatabaseURL:     getEnv("PAYMENT_DATABASE_URL", ""),
		RabbitMQURL:            getEnv("RABBITMQ_URL", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
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
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
