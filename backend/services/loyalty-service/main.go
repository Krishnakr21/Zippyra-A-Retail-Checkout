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
	cfg, err := config.Load("loyalty-service")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Initialize Database
	database, err := db.ConnectDB(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to Postgres: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize Kafka Producer & Consumer
	producer := kafka.NewProducer(cfg.Kafka.Brokers)
	repo := NewPostgresRepository(database.DB)
	consumer := NewEventConsumer(repo, producer)

	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, "loyalty-service-group", []string{"order.completed", "order.returned"})
	kafkaConsumer.RegisterHandler("order.completed", consumer.ProcessOrderCompleted)
	kafkaConsumer.RegisterHandler("order.returned", consumer.ProcessOrderReturned)

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go kafkaConsumer.Start(consumerCtx)

	customerHandler := NewCustomerHandler(repo)
	internalHandler := NewInternalHandler(repo)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RecoverMiddleware)

	RegisterRoutes(r, customerHandler, internalHandler, database.DB, cfg.JWT.Secret)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Loyalty Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Loyalty Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}
