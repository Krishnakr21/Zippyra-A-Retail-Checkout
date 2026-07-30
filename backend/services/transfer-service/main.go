package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/transfer_db?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}

	inventoryBaseURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryBaseURL == "" {
		inventoryBaseURL = "http://localhost:8088"
	}

	kafkaBrokers := "localhost:9092"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Error("Failed to connect to postgres: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	repo := NewPostgresRepository(db)
	inventoryClient := NewHTTPInventoryClient(inventoryBaseURL, jwtSecret)
	handler := NewTransferHandler(repo, inventoryClient, producer)
	router := NewRouter(handler, jwtSecret)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("transfer-service listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)
	logger.Info("transfer-service stopped gracefully")
}
