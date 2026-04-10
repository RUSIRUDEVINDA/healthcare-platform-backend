package main

import (
	"context"
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
	"healthcare-platform/services/ai-symptom-service/internal/service"
	"healthcare-platform/pkg/logger"
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

	svc := service.NewSymptomService(llm, log)
	h := handler.NewSymptomHandler(svc, log)

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
