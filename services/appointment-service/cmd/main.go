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
	"healthcare-platform/services/appointment-service/internal/config"
	"healthcare-platform/services/appointment-service/internal/handler"
	"healthcare-platform/services/appointment-service/internal/messaging"
	appmiddleware "healthcare-platform/services/appointment-service/internal/middleware"
	"healthcare-platform/services/appointment-service/internal/repository"
	"healthcare-platform/services/appointment-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load appointment-service config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to appointment database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run appointment migrations", "error", err)
	}

	jwtHelper := jwt.New(cfg.JWTSecret, cfg.JWTSecret, 15, 7)

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ from appointment-service", "error", err)
	}
	defer mqClient.Close()

	// Setup Business Logic
	appointmentRepo := repository.NewAppointmentRepository(db)
	appointmentSvc := service.NewAppointmentService(
		appointmentRepo,
		mqClient,
		log,
		cfg.DoctorServiceURL,
		cfg.PatientServiceURL,
		cfg.InternalAPIKey,
		cfg.JitsiBaseURL,
	)
	appointmentHandler := handler.NewAppointmentHandler(appointmentSvc, log)
	appointmentConsumer := messaging.NewAppointmentConsumer(mqClient, appointmentSvc, log)

	// Setup Gin HTTP Router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(appmiddleware.CORS())

	appointmentHandler.RegisterRoutes(router, appmiddleware.JWTAuth(jwtHelper))

	if err := appointmentConsumer.Start(); err != nil {
		log.Fatal("Failed to start payment completed consumer", "error", err)
	}

	stopCleaner := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := appointmentSvc.ExpireOverdueAppointments(time.Now().UTC()); err != nil {
					log.Warn("Failed to expire overdue appointments", "error", err)
				}
			case <-stopCleaner:
				return
			}
		}
	}()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("Appointment Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Appointment Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down appointment service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Appointment service forced shutdown", "error", err)
	}
	close(stopCleaner)
	log.Info("Appointment service exited gracefully")
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
	CREATE TABLE IF NOT EXISTS appointments (
		id               UUID PRIMARY KEY,
		patient_id       UUID NOT NULL,
		doctor_id        TEXT NOT NULL,
		doctor_owner_user_id TEXT NOT NULL DEFAULT '',
		slot_id          UUID,
		consultation_mode TEXT NOT NULL DEFAULT 'physical',
		room_name        TEXT NOT NULL DEFAULT '',
		join_url         TEXT NOT NULL DEFAULT '',
		scheduled_at     TIMESTAMPTZ NOT NULL,
		duration_minutes INT NOT NULL DEFAULT 30,
		status           TEXT NOT NULL DEFAULT 'pending',
		payment_status   TEXT NOT NULL DEFAULT 'pending',
		payment_due_at   TIMESTAMPTZ,
		paid_at          TIMESTAMPTZ,
		notes            TEXT,
		created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS slots (
		id         UUID PRIMARY KEY,
		doctor_id  TEXT NOT NULL,
		owner_user_id TEXT NOT NULL DEFAULT '',
		hospital   TEXT NOT NULL DEFAULT '',
		start_time TIMESTAMPTZ NOT NULL,
		end_time   TIMESTAMPTZ NOT NULL,
		is_booked  BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Backward compatible: older DBs used UUID for doctor_id; allow numeric IDs by converting to TEXT.
	ALTER TABLE appointments ALTER COLUMN doctor_id TYPE TEXT USING doctor_id::text;
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS doctor_owner_user_id TEXT NOT NULL DEFAULT '';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS slot_id UUID;
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS consultation_mode TEXT NOT NULL DEFAULT 'physical';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS room_name TEXT NOT NULL DEFAULT '';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS join_url TEXT NOT NULL DEFAULT '';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'pending';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS payment_due_at TIMESTAMPTZ;
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ;
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS patient_first_name TEXT NOT NULL DEFAULT '';
	ALTER TABLE appointments ADD COLUMN IF NOT EXISTS patient_last_name TEXT NOT NULL DEFAULT '';
	ALTER TABLE slots ALTER COLUMN doctor_id TYPE TEXT USING doctor_id::text;
	ALTER TABLE slots ADD COLUMN IF NOT EXISTS owner_user_id TEXT NOT NULL DEFAULT '';
	ALTER TABLE slots ADD COLUMN IF NOT EXISTS hospital TEXT NOT NULL DEFAULT '';

	CREATE INDEX IF NOT EXISTS idx_appointments_patient ON appointments(patient_id);
	CREATE INDEX IF NOT EXISTS idx_appointments_doctor  ON appointments(doctor_id);
	CREATE INDEX IF NOT EXISTS idx_appointments_doctor_owner ON appointments(doctor_owner_user_id);
	CREATE INDEX IF NOT EXISTS idx_appointments_slot ON appointments(slot_id);
	CREATE INDEX IF NOT EXISTS idx_slots_doctor         ON slots(doctor_id);
	CREATE INDEX IF NOT EXISTS idx_slots_owner          ON slots(owner_user_id);
	CREATE INDEX IF NOT EXISTS idx_slots_hospital       ON slots(hospital);
	CREATE INDEX IF NOT EXISTS idx_slots_start_time     ON slots(start_time);
	CREATE INDEX IF NOT EXISTS idx_appointments_payment_due ON appointments(payment_due_at);
	CREATE INDEX IF NOT EXISTS idx_appointments_payment_status ON appointments(payment_status);
	`
	_, err := db.Exec(migrationSQL)
	if err != nil {
		return err
	}
	log.Info("Appointment database migrations completed")
	return nil
}
