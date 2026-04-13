package main

import (
	"context"
	"database/sql"
	"errors"
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
	"healthcare-platform/services/file-storage-service/internal/config"
	"healthcare-platform/services/file-storage-service/internal/handler"
	"healthcare-platform/services/file-storage-service/internal/messaging"
	"healthcare-platform/services/file-storage-service/internal/middleware"
	"healthcare-platform/services/file-storage-service/internal/repository"
	"healthcare-platform/services/file-storage-service/internal/service"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load file-storage-service config", "error", err)
	}

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to connect to file storage database", "error", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run file storage migrations", "error", err)
	}

	jwtHelper := jwt.New(cfg.JWTSecret, cfg.JWTRefreshSecret, 15, 7)

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ from file-storage-service", "error", err)
	}
	defer mqClient.Close()

	cloudinaryProvider, err := service.NewCloudinaryProvider(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary provider", "error", err)
	}
	r2Provider, err := service.NewR2Provider(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Cloudflare R2 provider", "error", err)
	}

	fileRepo := repository.NewFileRepository(db)
	fileSvc := service.NewFileService(fileRepo, cloudinaryProvider, r2Provider, mqClient, log, cfg)
	fileHandler := handler.NewFileHandler(fileSvc, log)
	fileConsumer := messaging.NewFileEventConsumer(mqClient, fileSvc, log)

	if err := fileConsumer.Start(); err != nil {
		log.Fatal("Failed to start file storage consumer", "error", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(log))

	fileHandler.RegisterRoutes(router, middleware.RequireAuth(jwtHelper))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("File Storage Service started", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("File Storage Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down file storage service...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("File storage service forced shutdown", "error", err)
	}
	log.Info("File storage service exited gracefully")
}

func connectDB(url string, log *logger.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", url)
		if err == nil {
			err = db.Ping()
			if err == nil {
				return db, nil
			}
		}
		log.Warn("DB connection retry", "attempt", i, "error", err)
		time.Sleep(time.Duration(i) * 3 * time.Second)
	}
	return db, err
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id UUID PRIMARY KEY,
			owner_id TEXT NOT NULL,
			uploader_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			storage_provider TEXT NOT NULL,
			original_name TEXT NOT NULL,
			stored_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes BIGINT NOT NULL DEFAULT 0,
			checksum TEXT NOT NULL DEFAULT '',
			cloudinary_public_id TEXT NOT NULL DEFAULT '',
			cloudinary_url TEXT NOT NULL DEFAULT '',
			r2_bucket TEXT NOT NULL DEFAULT '',
			r2_object_key TEXT NOT NULL DEFAULT '',
			is_public BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		ALTER TABLE files ADD COLUMN IF NOT EXISTS r2_bucket TEXT NOT NULL DEFAULT '';
		ALTER TABLE files ADD COLUMN IF NOT EXISTS r2_object_key TEXT NOT NULL DEFAULT '';
		ALTER TABLE files DROP COLUMN IF EXISTS gcs_bucket;
		ALTER TABLE files DROP COLUMN IF EXISTS gcs_object_name;

		CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files(owner_id);
		CREATE INDEX IF NOT EXISTS idx_files_uploader_id ON files(uploader_id);
		CREATE INDEX IF NOT EXISTS idx_files_kind ON files(kind);
		CREATE INDEX IF NOT EXISTS idx_files_storage_provider ON files(storage_provider);
		CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
	`)
	return err
}
