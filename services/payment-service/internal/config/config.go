package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	RabbitMQURL string

	PayHereMerchantID     string
	PayHereMerchantSecret string
	PayHereEnv            string // "sandbox" or "live"
	PayHereReturnURL      string
	PayHereCancelURL      string
	PayHereNotifyURL      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8005"), // Different port for payment-service
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),

		PayHereMerchantID:     getEnv("PAYHERE_MERCHANT_ID", ""),
		PayHereMerchantSecret: getEnv("PAYHERE_MERCHANT_SECRET", ""),
		PayHereEnv:            getEnv("PAYHERE_ENV", "sandbox"),
		PayHereReturnURL:      getEnv("PAYHERE_RETURN_URL", "http://localhost:3000/payment/success"),
		PayHereCancelURL:      getEnv("PAYHERE_CANCEL_URL", "http://localhost:3000/payment/cancel"),
		PayHereNotifyURL:      getEnv("PAYHERE_NOTIFY_URL", "http://localhost:8005/api/v1/payments/webhook/payhere"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.RabbitMQURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL is not set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
