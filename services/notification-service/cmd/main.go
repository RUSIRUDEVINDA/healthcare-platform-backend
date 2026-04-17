package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/notification-service/internal/config"
	"healthcare-platform/services/notification-service/internal/handler"
	"healthcare-platform/services/notification-service/internal/messaging"
	"healthcare-platform/services/notification-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load notification-service config", "error", err)
	}

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", "error", err)
	}
	defer mqClient.Close()

	notificationSvc := service.NewNotificationService(cfg, log)
	notificationConsumer := messaging.NewConsumer(mqClient, notificationSvc, log)

	if err := mqClient.ConsumeQueue(
		"notification_appointment_booked_queue",
		rabbitmq.ExchangeAppointmentEvents,
		notificationConsumer.HandleAppointmentBooked,
		rabbitmq.RoutingKeyAppointmentBooked,
		"appointment_booked",
		"appointment-booked",
	); err != nil {
		log.Fatal("Failed to subscribe appointment.booked", "error", err)
	}

	if err := mqClient.ConsumeQueue(
		"notification_appointment_cancelled_queue",
		rabbitmq.ExchangeAppointmentEvents,
		notificationConsumer.HandleAppointmentCancelled,
		rabbitmq.RoutingKeyAppointmentCancelled,
		"appointment_cancelled",
		"appointment-cancelled",
	); err != nil {
		log.Fatal("Failed to subscribe appointment.cancelled", "error", err)
	}

	if err := mqClient.ConsumeQueue(
		"notification_payment_completed_queue",
		rabbitmq.ExchangePaymentEvents,
		notificationConsumer.HandlePaymentCompleted,
		rabbitmq.RoutingKeyPaymentCompleted,
		"payment_completed",
		"payment-completed",
	); err != nil {
		log.Fatal("Failed to subscribe payment.completed", "error", err)
	}

	if err := mqClient.ConsumeQueue(
		"notification_consultation_completed_queue",
		rabbitmq.ExchangeAppointmentEvents,
		notificationConsumer.HandleConsultationCompleted,
		rabbitmq.RoutingKeyConsultationCompleted,
		"consultation_completed",
		"consultation-completed",
	); err != nil {
		log.Fatal("Failed to subscribe consultation.completed", "error", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	handler.NewHealthHandler().RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Notification Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Notification server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down notification service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Notification service forced shutdown", "error", err)
	}

	log.Info("Notification service exited gracefully")
}
