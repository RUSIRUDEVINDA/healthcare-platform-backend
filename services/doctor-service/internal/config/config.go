package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// Config holds environment-driven settings (aligned with auth-service pattern).
type Config struct {
	Port            string
	AppEnv          string
	DatabaseURL     string
	RabbitMQURL     string
	AuthServiceURL  string
}

// Load reads configuration from the environment (and optional .env file).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8003"),
		AppEnv:         getEnv("APP_ENV", "development"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RabbitMQURL:    getEnv("RABBITMQ_URL", ""),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
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
	if c.AuthServiceURL == "" {
		return errors.New("AUTH_SERVICE_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
