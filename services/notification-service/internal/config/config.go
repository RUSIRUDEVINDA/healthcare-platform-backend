package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	Port        string
	RabbitMQURL string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		Port:             getEnv("PORT", "8006"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnvInt("SMTP_PORT", 587),
		SMTPUsername:     getEnv("SMTP_USERNAME", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		FromEmail:        getEnv("FROM_EMAIL", ""),
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
