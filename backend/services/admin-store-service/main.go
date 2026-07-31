package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/admin_store_db?sslmode=disable"
	}

	storeServiceURL := os.Getenv("STORE_SERVICE_URL")
	if storeServiceURL == "" {
		storeServiceURL = "http://localhost:8010"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	// Connect to DB; fall back to memory repo when DB is unavailable (dev / test)
	var repo ChainRepository
	db, err := sql.Open("postgres", dbURL)
	if err == nil {
		if pingErr := db.PingContext(context.Background()); pingErr == nil {
			repo = NewPostgresChainRepository(db)
			logger.Info("admin-store-service connected to PostgreSQL")
		} else {
			logger.Info("admin-store-service: DB ping failed (%v) — using in-memory fallback", pingErr)
			repo = NewMemoryChainRepository()
		}
	} else {
		logger.Info("admin-store-service: DB open failed — using in-memory fallback")
		repo = NewMemoryChainRepository()
	}

	storeClient := NewStoreServiceClient(storeServiceURL, jwtSecret)
	chainHandler := NewChainHandler(repo, jwtSecret)
	storeHandler := NewStoreAdminHandler(repo, storeClient, jwtSecret)

	router := SetupRoutes(chainHandler, storeHandler)

	serverAddr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("admin-store-service running on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down admin-store-service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown: %v", err)
	} else {
		logger.Info("admin-store-service stopped cleanly")
	}
}
