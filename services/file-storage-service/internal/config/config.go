package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	Port                   string
	DatabaseURL            string
	RabbitMQURL            string
	AppointmentServiceURL  string
	JWTSecret              string
	JWTRefreshSecret       string
	CloudinaryCloudName    string
	CloudinaryAPIKey       string
	CloudinaryAPISecret    string
	CloudinaryUploadFolder string
	R2AccountID            string
	R2AccessKeyID          string
	R2SecretAccessKey      string
	R2Bucket               string
	R2Endpoint             string
	MaxImageSizeBytes      int64
	MaxDocumentSizeBytes   int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		Port:                   getEnv("PORT", "8010"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		RabbitMQURL:            getEnv("RABBITMQ_URL", ""),
		AppointmentServiceURL:  getEnv("APPOINTMENT_SERVICE_URL", "http://localhost:8004"),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:       getEnv("JWT_REFRESH_SECRET", ""),
		CloudinaryCloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:       getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryUploadFolder: getEnv("CLOUDINARY_UPLOAD_FOLDER", "healthcare-platform"),
		R2AccountID:            getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:          getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:      getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:               getEnv("R2_BUCKET", ""),
		R2Endpoint:             getEnv("R2_ENDPOINT", ""),
		MaxImageSizeBytes:      getEnvInt64("MAX_IMAGE_SIZE_BYTES", 10*1024*1024),
		MaxDocumentSizeBytes:   getEnvInt64("MAX_DOCUMENT_SIZE_BYTES", 25*1024*1024),
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
	if cfg.JWTRefreshSecret == "" {
		cfg.JWTRefreshSecret = cfg.JWTSecret
	}
	if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("cloudinary configuration is incomplete")
	}
	if cfg.R2AccountID == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2Bucket == "" {
		return nil, fmt.Errorf("cloudflare r2 configuration is incomplete")
	}
	if cfg.R2Endpoint == "" {
		cfg.R2Endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}
