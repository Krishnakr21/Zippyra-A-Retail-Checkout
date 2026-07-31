package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	logger.Info("Starting subscription-service...")

	cfg, err := config.Load("subscription-service")
	var database *db.DB
	if err == nil {
		database, _ = db.ConnectDB(cfg.Database.URL)
	}

	var sqlDB = (*db.DB)(nil)
	var repo Repository
	if database != nil {
		sqlDB = database
		repo = NewPostgresRepository(database.DB)
		defer database.Close()
	} else {
		repo = NewPostgresRepository(nil)
	}

	_ = sqlDB
	jwtSecret := os.Getenv("JWT_SECRET")
	handler := NewSubscriptionHandler(repo, jwtSecret)
	routes := SetupRoutes(handler)

	port := os.Getenv("SUBSCRIPTION_SERVICE_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8104"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      routes,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("subscription-service listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down subscription-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("subscription-service stopped cleanly.")
}
