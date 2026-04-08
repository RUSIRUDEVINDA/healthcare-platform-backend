package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds environment-driven settings. Stateless service — no database.
type Config struct {
	Port           string
	AppEnv         string
	AuthServiceURL string
	AIProvider     string // openai | gemini

	OpenAIAPIKey  string
	OpenAIModel   string
	GeminiAPIKey  string
	GeminiModel   string
}

// Load reads configuration from the environment (and optional .env file).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8008"),
		AppEnv:         getEnv("APP_ENV", "development"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		AIProvider:     strings.ToLower(strings.TrimSpace(getEnv("AI_PROVIDER", "openai"))),
		OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		GeminiModel:    getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.AuthServiceURL == "" {
		return errors.New("AUTH_SERVICE_URL is required")
	}
	switch c.AIProvider {
	case "openai":
		if c.OpenAIAPIKey == "" {
			return errors.New("OPENAI_API_KEY is required when AI_PROVIDER=openai")
		}
	case "gemini":
		if c.GeminiAPIKey == "" {
			return errors.New("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
		}
	default:
		return errors.New("AI_PROVIDER must be openai or gemini")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
