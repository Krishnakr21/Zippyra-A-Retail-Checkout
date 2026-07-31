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
	cfg, err := config.Load("integration-service")
	if err != nil {
		logger.Info("Config load warning (using defaults): %v", err)
	}

	masterKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	if masterKey == "" {
		masterKey = "zippyra-dev-totp-encryption-key32"
	}

	// 2. Connect Database
	var repo IntegrationRepository
	if cfg != nil && cfg.Database.Host != "" {
		dsn := os.Getenv("DATABASE_URL")
		database, err := db.ConnectDB(dsn)
		if err != nil {
			logger.Warn("Failed to connect to Postgres DB, using in-memory repo: %v", err)
			repo = NewMemoryIntegrationRepository()
		} else {
			repo = NewPostgresIntegrationRepository(database.DB)
			logger.Info("Connected to PostgreSQL database for integration-service")
		}
	} else {
		logger.Info("Using in-memory repository fallback for integration-service")
		repo = NewMemoryIntegrationRepository()
	}

	// 3. Initialize components
	directPushWorker := NewDirectPushWorker(repo, masterKey)
	connHandler := NewConnectionHandler(repo, nil, masterKey)
	webhookHandler := NewWebhookHandler(repo, masterKey, nil)
	agentHandler := NewAgentHandler(repo)
	syncJobHandler := NewSyncJobHandler(repo, directPushWorker)

	// 4. Setup Routes
	r := mux.NewRouter()
	SetupRoutes(r, connHandler, webhookHandler, agentHandler, syncJobHandler, nil)

	// 5. Start Background Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go directPushWorker.StartBackgroundRetryLoop(ctx, 5*time.Minute)

	// 6. Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Integration Service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Integration Service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}
	logger.Info("Integration Service exited cleanly")
}
