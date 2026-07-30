package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	// 1. Load config
	cfg, err := config.Load("notification-service")
	if err != nil {
		logger.Info("Config load warning (using defaults): %v", err)
	}

	// 2. Connect Database
	var repo NotificationRepository
	if cfg != nil && cfg.Database.Host != "" {
		dsn := os.Getenv("DATABASE_URL")
		database, err := db.ConnectDB(dsn)
		if err != nil {
			logger.Warn("Failed to connect to Postgres DB, using in-memory repo: %v", err)
			repo = NewMemoryRepository()
		} else {
			repo = NewPostgresRepository(database.DB)
			logger.Info("Connected to PostgreSQL database for notification-service")
		}
	} else {
		logger.Info("Using in-memory repository fallback for notification-service")
		repo = NewMemoryRepository()
	}

	// 3. Initialize Clients & Engine
	fcmClient := NewMockFCMClient()
	whatsAppClient := NewMockWhatsAppClient()
	opsAlerts := NewOpsAlertDispatcher(repo)

	engine := NewNotificationEngine(repo, fcmClient, whatsAppClient)
	_ = NewEventConsumer(engine, opsAlerts)

	handler := NewNotificationHandler(repo)
	adminHandler := NewAdminHandler(repo)

	// 4. Setup Routes
	r := mux.NewRouter()
	SetupRoutes(r, handler, adminHandler, nil)

	// 5. Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Notification Service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown with 25s bounded context
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Notification Service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}
	logger.Info("Notification Service exited cleanly")
}
