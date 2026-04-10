package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/auth-service/internal/config"
	"healthcare-platform/services/auth-service/internal/handler"
	"healthcare-platform/services/auth-service/internal/middleware"
	"healthcare-platform/services/auth-service/internal/repository"
	"healthcare-platform/services/auth-service/internal/service"
	"healthcare-platform/pkg/jwt"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize structured logger
	log := logger.New()

	// Load config from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	// Connect to PostgreSQL with retry
	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Run database migrations
	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	// Connect to RabbitMQ with retry
	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", "error", err)
	}
	defer mqClient.Close()

	// Ensure patient-service queue is bound before we publish events.
	// This prevents losing user.registered events when patient-service starts later.
	if err := mqClient.EnsureQueueBindings(
		"patient_profile_creator_queue",
		rabbitmq.ExchangeUserEvents,
		rabbitmq.RoutingKeyUserRegistered,
		rabbitmq.RoutingKeyUserLoggedIn,
	); err != nil {
		log.Fatal("Failed to ensure RabbitMQ queue bindings", "error", err)
	}

	// Initialize JWT helper
	jwtHelper := jwt.New(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.AccessTokenTTLMinutes, cfg.RefreshTokenTTLDays)

	// Wire up layers: repository -> service -> handler
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtHelper, mqClient, log)
	authHandler := handler.NewAuthHandler(authSvc, log)

	// Setup Gin router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	// Register routes
	authHandler.RegisterRoutes(router)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/logout", authHandler.Logout)
	router.POST("/api/auth/refresh", authHandler.Refresh)
	router.GET("/api/auth/validate", authHandler.ValidateToken)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Auth Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

	// Graceful shutdown — wait for SIGINT or SIGTERM signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			log.Warn("DB ping failed, retrying...", "attempt", i, "error", err)
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info("Connected to PostgreSQL successfully")
	return db, nil
}

func runMigrations(db *sql.DB, log *logger.Logger) error {
	migrationSQL, err := os.ReadFile("migrations/0001_init.up.sql")
	if err != nil {
		// If file not found, use embedded SQL
		log.Warn("Migration file not found, using embedded SQL")
		migrationSQL = []byte(embeddedMigration)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return err
	}

	log.Info("Database migrations completed")
	return nil
}

// Fallback migration if file not found
const embeddedMigration = `
CREATE TABLE IF NOT EXISTS users (
	id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email         VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	role          VARCHAR(20) NOT NULL CHECK (role IN ('patient', 'doctor', 'admin')),
	first_name    VARCHAR(100) NOT NULL,
	last_name     VARCHAR(100) NOT NULL,
	is_verified   BOOLEAN DEFAULT FALSE,
	is_active     BOOLEAN DEFAULT TRUE,
	created_at    TIMESTAMPTZ DEFAULT NOW(),
	updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash VARCHAR(255) UNIQUE NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);
`
