package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port   string
	AppEnv string

	// Database
	DatabaseURL string

	// RabbitMQ
	RabbitMQURL string

	// JWT
	JWTSecret             string
	JWTRefreshSecret      string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
}

func Load() (*Config, error) {
	// Attempt to load .env file for local development
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8001"),
		AppEnv:                getEnv("APP_ENV", "development"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		RabbitMQURL:           getEnv("RABBITMQ_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:      getEnv("JWT_REFRESH_SECRET", ""),
		AccessTokenTTLMinutes: getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15),
		RefreshTokenTTLDays:   getEnvInt("REFRESH_TOKEN_TTL_DAYS", 7),
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
	if c.JWTRefreshSecret == "" {
		return errors.New("JWT_REFRESH_SECRET is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}
