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
	sharedmiddleware "healthcare-platform/pkg/middleware"
	"healthcare-platform/pkg/rabbitmq"
	tmconfig "healthcare-platform/services/telemedicine-service/internal/config"
	tmhandler "healthcare-platform/services/telemedicine-service/internal/handler"
	tmmessaging "healthcare-platform/services/telemedicine-service/internal/messaging"
	tmmiddleware "healthcare-platform/services/telemedicine-service/internal/middleware"
	"healthcare-platform/services/telemedicine-service/internal/repository"
	"healthcare-platform/services/telemedicine-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := tmconfig.Load()
	if err != nil {
		log.Fatal("Failed to load telemedicine-service config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to telemedicine database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run telemedicine migrations", "error", err)
	}

	jwtHelper := jwt.New(cfg.JWTSecret, cfg.JWTSecret, 15, 7)

	sessionRepo := repository.NewSessionRepository(db)
	sessionSvc := service.NewSessionService(sessionRepo, log, cfg.AppointmentServiceURL, cfg.JitsiBaseURL)
	sessionHandler := tmhandler.NewSessionHandler(sessionSvc, log)

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ from telemedicine-service", "error", err)
	}
	defer mqClient.Close()

	appointmentBookedConsumer := tmmessaging.NewAppointmentBookedConsumer(mqClient, sessionSvc, log)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sharedmiddleware.CORSMiddleware())

	sessionHandler.RegisterRoutes(router, tmmiddleware.JWTAuth(jwtHelper))

	if err := appointmentBookedConsumer.Start(); err != nil {
		log.Fatal("Failed to start appointment booked consumer", "error", err)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("Telemedicine Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Telemedicine server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down telemedicine service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Telemedicine service forced shutdown", "error", err)
	}

	log.Info("Telemedicine service exited gracefully")
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
	sqlBytes, err := os.ReadFile("migrations/001_telemedicine.sql")
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return err
	}
	log.Info("Telemedicine database migrations completed")
	return nil
}
