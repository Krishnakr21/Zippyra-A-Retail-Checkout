package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/middleware"
)

func main() {
	cfg, err := config.Load("warehouse-service")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	database, err := db.ConnectDB(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to Postgres: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	producer := kafka.NewProducer(cfg.Kafka.Brokers)
	repo := NewPostgresRepository(database.DB)

	inventoryBaseURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryBaseURL == "" {
		inventoryBaseURL = "http://localhost:8088"
	}
	inventoryClient := NewHTTPInventoryClient(inventoryBaseURL, cfg.JWT.Secret)

	// Setup Kafka Consumer for auto-reorder
	consumer := NewEventConsumer(repo, producer)
	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, "warehouse-service-group", []string{TopicLowStock})
	kafkaConsumer.RegisterHandler(TopicLowStock, consumer.ProcessLowStockEvent)

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go kafkaConsumer.Start(consumerCtx)

	qcBaseURL := os.Getenv("QC_SERVICE_URL")
	if qcBaseURL == "" {
		qcBaseURL = "http://localhost:8089"
	}
	qcClient := NewQCClient(qcBaseURL, cfg.JWT.Secret)

	transferBaseURL := os.Getenv("TRANSFER_SERVICE_URL")
	if transferBaseURL == "" {
		transferBaseURL = "http://localhost:8090"
	}
	transferClient := NewTransferClient(transferBaseURL, cfg.JWT.Secret)

	poHandler := NewPOHandler(repo)
	grnHandler := NewGRNHandler(repo, inventoryClient, qcClient, producer)
	transferHandler := NewTransferHandlerWithClient(repo, inventoryClient, transferClient, producer)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RecoverMiddleware)

	RegisterRoutes(r, poHandler, grnHandler, transferHandler, database.DB, cfg.JWT.Secret)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Warehouse Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Warehouse Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}
