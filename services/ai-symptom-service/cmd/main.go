package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/ai-symptom-service/internal/config"
	"healthcare-platform/services/ai-symptom-service/internal/handler"
	"healthcare-platform/services/ai-symptom-service/internal/integrations"
	"healthcare-platform/services/ai-symptom-service/internal/middleware"
	"healthcare-platform/services/ai-symptom-service/internal/repository"
	"healthcare-platform/services/ai-symptom-service/internal/service"
	"healthcare-platform/pkg/logger"

	_ "github.com/lib/pq"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	var llm integrations.LLMClient
	switch strings.ToLower(cfg.AIProvider) {
	case "openai":
		llm = integrations.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		log.Info("LLM provider", "provider", "openai", "model", cfg.OpenAIModel)
	case "gemini":
		llm = integrations.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
		log.Info("LLM provider", "provider", "gemini", "model", cfg.GeminiModel)
	default:
		log.Fatal("Invalid AI_PROVIDER", "value", cfg.AIProvider)
	}

	var chatRepo *repository.ChatRepository
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		db, err := connectDB(cfg.DatabaseURL, log)
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL", "error", err)
		}
		defer db.Close()
		if err := runMigrations(db, log); err != nil {
			log.Fatal("Failed to run migrations", "error", err)
		}
		chatRepo = repository.NewChatRepository(db)
		log.Info("Chat history persistence enabled")
	} else {
		log.Warn("DATABASE_URL not set; symptom chat history disabled")
	}

	svc := service.NewSymptomService(llm, log)
	h := handler.NewSymptomHandler(svc, chatRepo, log)

	authHTTP := &http.Client{Timeout: 15 * time.Second}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	h.RegisterRoutes(router, authHTTP, cfg.AuthServiceURL)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 95 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("AI Symptom Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited gracefully")
}

func connectDB(url string, log *logger.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", url)
		if err != nil {
			log.Warn("DB open failed, retrying...", "attempt", i, "error", err)
			time.Sleep(time.Duration(i) * 2 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			_ = db.Close()
			log.Warn("DB ping failed, retrying...", "attempt", i, "error", err)
			time.Sleep(time.Duration(i) * 2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Info("Connected to PostgreSQL successfully")
	return db, nil
}

func runMigrations(db *sql.DB, log *logger.Logger) error {
	files := []struct {
		path     string
		fallback string
	}{
		{"migrations/0001_symptom_chat_history.up.sql", embeddedSymptomChatMigration},
	}
	for _, f := range files {
		sqlBytes, err := os.ReadFile(f.path)
		if err != nil {
			log.Warn("Migration file not found, using embedded SQL", "file", f.path, "error", err)
			sqlBytes = []byte(f.fallback)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s: %w", f.path, err)
		}
	}
	log.Info("Database migrations completed")
	return nil
}

const embeddedSymptomChatMigration = `
CREATE TABLE IF NOT EXISTS symptom_chat_history (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             TEXT NOT NULL,
    symptoms            TEXT NOT NULL,
    optional_context    TEXT NOT NULL DEFAULT '',
    suggested_specialty TEXT NOT NULL,
    preliminary_notes   TEXT NOT NULL,
    disclaimer          TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_symptom_chat_user_created ON symptom_chat_history (user_id, created_at DESC);
`
