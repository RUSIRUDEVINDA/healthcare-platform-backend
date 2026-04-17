package config

import (
	"errors"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	Port        string
	RabbitMQURL string

	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	FromEmail      string
	SMTPSecure     bool
	SMTPSkipVerify bool

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
		SMTPUsername:     getEnvFallback("SMTP_USERNAME", "SMTP_USER", ""),
		SMTPPassword:     getEnvFallback("SMTP_PASSWORD", "SMTP_PASS", ""),
		FromEmail:        normalizeFromEmail(getEnvFallback("FROM_EMAIL", "EMAIL_FROM", ""), getEnvFallback("SMTP_USERNAME", "SMTP_USER", "")),
		SMTPSecure:       getEnvBool("SMTP_SECURE", false),
		SMTPSkipVerify:   getEnvBool("SMTP_SKIP_VERIFY", false),
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

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFallback(primary, alias, fallback string) string {
	if value, exists := os.LookupEnv(primary); exists && strings.TrimSpace(value) != "" {
		return value
	}
	if value, exists := os.LookupEnv(alias); exists && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func normalizeFromEmail(fromEmail, fallback string) string {
	fromEmail = strings.TrimSpace(fromEmail)
	if fromEmail == "" {
		return strings.TrimSpace(fallback)
	}
	if parsed, err := mail.ParseAddress(fromEmail); err == nil && parsed.Address != "" {
		return parsed.Address
	}
	if strings.Contains(fromEmail, "@") {
		return fromEmail
	}
	return strings.TrimSpace(fallback)
}
