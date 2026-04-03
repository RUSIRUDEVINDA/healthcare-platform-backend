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

	"healthcare-platform/services/patient-service/internal/config"
	"healthcare-platform/services/patient-service/internal/handler"
	"healthcare-platform/services/patient-service/internal/messaging"
	"healthcare-platform/services/patient-service/internal/repository"
	"healthcare-platform/services/patient-service/internal/service"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load patient-service config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to patient database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run patient migrations", "error", err)
	}

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ from patient-service", "error", err)
	}
	defer mqClient.Close()

	// Setup Business Logic
	patientRepo := repository.NewPatientRepository(db)
	patientSvc := service.NewPatientService(patientRepo, log)
	patientHandler := handler.NewPatientHandler(patientSvc, log)

	// Setup Messaging Consumer
	// This listens for user registration events to create profiles
	patientConsumer := messaging.NewPatientConsumer(mqClient, patientSvc, log)
	if err := patientConsumer.Start(); err != nil {
		log.Fatal("Failed to start patient consumer", "error", err)
	}

	// Setup Gin HTTP Router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	// router.Use(middleware.Logger(log)) // If we had one for this service

	patientHandler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("Patient Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Patient Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down patient service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Patient service forced shutdown", "error", err)
	}
	log.Info("Patient service exited gracefully")
}

func connectDB(url string, log *logger.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", url)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Warn("DB connection retry", "attempt", i, "error", err)
		time.Sleep(time.Duration(i) * 3 * time.Second)
	}
	return db, err
}

func runMigrations(db *sql.DB, log *logger.Logger) error {
	migrationSQL := `
	CREATE TABLE IF NOT EXISTS patients (
		id                 UUID PRIMARY KEY,
		user_id            UUID UNIQUE NOT NULL,
		email              VARCHAR(255) NOT NULL,
		first_name         VARCHAR(100) NOT NULL,
		last_name          VARCHAR(100) NOT NULL,
		date_of_birth      DATE,
		gender             VARCHAR(20),
		phone_number       VARCHAR(20),
		address            TEXT,
		emergency_contact  VARCHAR(255),
		blood_group        VARCHAR(5),
		created_at         TIMESTAMPTZ DEFAULT NOW(),
		updated_at         TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_patients_user_id ON patients(user_id);
	`
	_, err := db.Exec(migrationSQL)
	if err != nil {
		return err
	}
	log.Info("Patient database migrations completed")
	return nil
}
