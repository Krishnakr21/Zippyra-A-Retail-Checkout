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
	cfg, err := config.Load("support-service")
	if err != nil {
		logger.Info("Config load warning (using defaults): %v", err)
	}

	var repo SupportRepository
	if cfg != nil && cfg.Database.Host != "" {
		dsn := os.Getenv("DATABASE_URL")
		database, err := db.ConnectDB(dsn)
		if err != nil {
			logger.Warn("Failed to connect to Postgres DB, using in-memory repo: %v", err)
			repo = NewMemoryRepository()
		} else {
			repo = NewPostgresRepository(database.DB)
			logger.Info("Connected to PostgreSQL database for support-service")
		}
	} else {
		logger.Info("Using in-memory repository fallback for support-service")
		repo = NewMemoryRepository()
	}

	handler := NewTicketHandler(repo, nil, nil)

	// SLA Warning Job (runs every 30 minutes)
	slaJob := NewSLAWarningJob(repo, nil)
	slaJob.Start(30 * time.Minute)

	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	port := os.Getenv("SUPPORT_SERVICE_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8093"
	}
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Support Service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Support Service...")
	slaJob.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}
	logger.Info("Support Service exited cleanly")
}
