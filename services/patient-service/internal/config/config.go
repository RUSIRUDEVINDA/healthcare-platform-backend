package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	Port            string
	DatabaseURL     string
	RabbitMQURL     string
	JWTSecret       string
	InternalAPIKey  string // optional; enables /internal/v1 routes when set
}

func Load() (*Config, error) {
	// Try loading .env from current dir for local dev
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8002"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RabbitMQURL:    getEnv("RABBITMQ_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		InternalAPIKey: getEnv("INTERNAL_API_KEY", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.RabbitMQURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL is not set")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
