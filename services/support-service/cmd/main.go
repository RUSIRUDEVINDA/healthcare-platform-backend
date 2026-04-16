package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"healthcare-platform/services/support-service/internal/config"
	"healthcare-platform/services/support-service/internal/handler"
	"healthcare-platform/services/support-service/internal/repository"
	"healthcare-platform/services/support-service/internal/service"
)

func main() {
	cfg := config.Load()

	// PostgreSQL connection with retry
	db := connectDB(cfg.DatabaseURL)
	defer db.Close()

	// Dependency injection
	repo := repository.NewTicketRepository(db)

	// Auto-migrate schema
	if err := repo.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	svc := service.NewTicketService(repo)
	h := handler.NewTicketHandler(svc)

	router := gin.Default()

	// Routes
	api := router.Group("/api/support")
	{
		api.POST("/tickets", h.Create)
		api.GET("/tickets", h.List) // Internal/Admin
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Support Service listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Support Service...")
	log.Println("Support Service exited")
}

func connectDB(url string) *sql.DB {
	var db *sql.DB
	var err error

	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", url)
		if err != nil {
			log.Printf("DB open failed (attempt %d): %v", i, err)
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			log.Printf("DB ping failed (attempt %d): %v", i, err)
			time.Sleep(time.Duration(i) * 3 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Println("Connected to PostgreSQL (support-db)")
	return db
}
