package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/middleware"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/payment-service/internal/config"
	"healthcare-platform/services/payment-service/internal/handler"
	"healthcare-platform/services/payment-service/internal/messaging"
	"healthcare-platform/services/payment-service/internal/provider"
	"healthcare-platform/services/payment-service/internal/repository"
	"healthcare-platform/services/payment-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load payment-service config", "error", err)
	}
	if strings.Contains(cfg.PayHereNotifyURL, "localhost") || strings.Contains(cfg.PayHereNotifyURL, "127.0.0.1") {
		log.Warn("PAYHERE_NOTIFY_URL points to localhost; PayHere cannot reach it. Use a public URL (e.g., ngrok).", "notify_url", cfg.PayHereNotifyURL)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to payment database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db, log); err != nil {
		log.Fatal("Failed to run payment migrations", "error", err)
	}

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ from payment-service", "error", err)
	}
	defer mqClient.Close()

	// Initialize Layers
	repo := repository.NewPaymentRepository(db)
	paymentProvider := provider.NewPayHereProvider(
		cfg.PayHereMerchantID,
		cfg.PayHereMerchantSecret,
		cfg.PayHereEnv,
		cfg.PayHereReturnURL,
		cfg.PayHereCancelURL,
		cfg.PayHereNotifyURL,
	)
	eventPublisher := messaging.NewRabbitMQEventPublisher(mqClient)
	svc := service.NewPaymentService(repo, paymentProvider, eventPublisher, log)
	h := handler.NewPaymentHandler(svc, log)

	// Messaging
	consumer := messaging.NewPaymentConsumer(mqClient, svc, log)
	if err := consumer.Start(); err != nil {
		log.Fatal("Failed to start payment consumer", "error", err)
	}

	// Router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})


	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info("Payment Service started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Payment Server failed", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down payment service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Payment service forced shutdown", "error", err)
	}
	log.Info("Payment service exited gracefully")
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
	CREATE TABLE IF NOT EXISTS payments (
		id             UUID PRIMARY KEY,
		appointment_id UUID UNIQUE NOT NULL,
		patient_id     UUID NOT NULL,
		amount         DECIMAL(10,2) NOT NULL,
		currency       VARCHAR(3) NOT NULL,
		status         VARCHAR(20) NOT NULL,
		provider       VARCHAR(20) NOT NULL,
		provider_id    VARCHAR(255),
		created_at     TIMESTAMPTZ DEFAULT NOW(),
		updated_at     TIMESTAMPTZ DEFAULT NOW()
	);
	-- If the table already existed from a previous version, ensure new columns exist.
	-- (CREATE TABLE IF NOT EXISTS does not add missing columns.)
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS provider       VARCHAR(20)  NOT NULL DEFAULT 'payhere';
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS provider_id    VARCHAR(255);
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS status         VARCHAR(20)  NOT NULL DEFAULT 'pending';
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS currency       VARCHAR(3)   NOT NULL DEFAULT 'USD';
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS created_at     TIMESTAMPTZ  DEFAULT NOW();
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS updated_at     TIMESTAMPTZ  DEFAULT NOW();
	CREATE INDEX IF NOT EXISTS idx_payments_appointment_id ON payments(appointment_id);
	CREATE INDEX IF NOT EXISTS idx_payments_patient_id ON payments(patient_id);
	`
	_, err := db.Exec(migrationSQL)
	if err != nil {
		return err
	}
	log.Info("Payment database migrations completed")
	return nil
}
