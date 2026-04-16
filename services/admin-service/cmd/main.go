package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"healthcare-platform/pkg/jwt"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/admin-service/internal/config"
	"healthcare-platform/services/admin-service/internal/handler"
	"healthcare-platform/services/admin-service/internal/messaging"
	"healthcare-platform/services/admin-service/internal/repository"
	"healthcare-platform/services/admin-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load admin-service config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to admin database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run admin migrations", "error", err)
	}

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", "error", err)
	}
	defer mqClient.Close()

	jwtHelper := jwt.New(cfg.JWTSecret, "", 0, 0)

	repo := repository.NewAdminRepository(db)
	adminSvc := service.NewAdminService(repo, log)
	consumer := messaging.NewConsumer(mqClient, adminSvc, log)
	if err := consumer.Start(); err != nil {
		log.Fatal("Failed to start admin consumers", "error", err)
	}

	adminHandler := handler.NewAdminHandler(adminSvc, jwtHelper, log)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	adminHandler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Admin Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Admin server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down admin service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Admin service forced shutdown", "error", err)
	}

	log.Info("Admin service exited gracefully")
}

func connectDB(url string, log *logger.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", url)
		if err != nil {
			log.Warn("Admin DB open failed, retrying...", "attempt", i, "error", err)
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			log.Warn("Admin DB ping failed, retrying...", "attempt", i, "error", err)
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

	log.Info("Connected to admin PostgreSQL successfully")
	return db, nil
}

func runMigrations(db *sql.DB, log *logger.Logger) error {
	migrationSQL, err := os.ReadFile("migrations/0001_init.up.sql")
	if err != nil {
		log.Warn("Admin migration file not found, using embedded SQL")
		migrationSQL = []byte(adminEmbeddedMigration)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return err
	}

	log.Info("Admin database migrations completed")
	return nil
}

const adminEmbeddedMigration = `
CREATE TABLE IF NOT EXISTS admin_users (
	id UUID PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	role VARCHAR(20) NOT NULL,
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	is_verified BOOLEAN DEFAULT FALSE,
	is_active BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS doctor_verifications (
	doctor_id UUID PRIMARY KEY REFERENCES admin_users(id) ON DELETE CASCADE,
	verified_by UUID,
	notes TEXT,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	verified_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS appointments (
	id UUID PRIMARY KEY,
	patient_id UUID NOT NULL,
	doctor_id UUID NOT NULL,
	status VARCHAR(30) NOT NULL,
	scheduled_at TIMESTAMPTZ NOT NULL,
	reason TEXT,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	amount NUMERIC(12,2) NOT NULL,
	currency VARCHAR(10) NOT NULL,
	status VARCHAR(30) NOT NULL,
	provider VARCHAR(50) NOT NULL,
	reference VARCHAR(100) UNIQUE,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_doctor_verifications_status ON doctor_verifications(status);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
`
