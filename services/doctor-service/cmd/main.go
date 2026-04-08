package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/services/doctor-service/internal/config"
	"healthcare-platform/services/doctor-service/internal/handler"
	"healthcare-platform/services/doctor-service/internal/middleware"
	"healthcare-platform/services/doctor-service/internal/repository"
	"healthcare-platform/services/doctor-service/internal/service"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"

	_ "github.com/lib/pq"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	var mqClient *rabbitmq.Client
	if cfg.RabbitMQURL != "" {
		var mqErr error
		mqClient, mqErr = rabbitmq.NewClient(cfg.RabbitMQURL, log)
		if mqErr != nil {
			log.Fatal("Failed to connect to RabbitMQ", "error", mqErr)
		}
		defer mqClient.Close()
	} else {
		log.Warn("RABBITMQ_URL not set; doctor events will not be published")
	}

	docRepo := repository.NewDoctorRepository(db)
	docSvc := service.NewDoctorService(docRepo, mqClient, log)
	docHandler := handler.NewDoctorHandler(docSvc, log)

	authHTTP := &http.Client{Timeout: 10 * time.Second}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())

	docHandler.RegisterRoutes(router, authHTTP, cfg.AuthServiceURL)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Doctor Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

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
	files := []struct {
		path     string
		fallback string
	}{
		{"migrations/0001_doctors.up.sql", embeddedDoctorsMigration},
		{"migrations/0002_doctors_nic_slmc.up.sql", embeddedDoctorsNicSlmcMigration},
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

const embeddedDoctorsMigration = `
CREATE TABLE IF NOT EXISTS doctors (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    specialization  VARCHAR(255) NOT NULL,
    experience      INT NOT NULL CHECK (experience >= 0 AND experience <= 80),
    hospital        VARCHAR(255) NOT NULL,
    nic             VARCHAR(12) NOT NULL,
    slmc_no         VARCHAR(5) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_doctors_specialization ON doctors (specialization);
`

const embeddedDoctorsNicSlmcMigration = `
ALTER TABLE doctors ADD COLUMN IF NOT EXISTS nic VARCHAR(12);
ALTER TABLE doctors ADD COLUMN IF NOT EXISTS slmc_no VARCHAR(5);
UPDATE doctors SET
    nic = RIGHT(REPEAT('0', 12) || id::text, 12),
    slmc_no = RIGHT(REPEAT('0', 5) || id::text, 5)
WHERE nic IS NULL OR slmc_no IS NULL;
ALTER TABLE doctors ALTER COLUMN nic SET NOT NULL;
ALTER TABLE doctors ALTER COLUMN slmc_no SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_nic ON doctors (nic);
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_slmc_no ON doctors (slmc_no);
`
