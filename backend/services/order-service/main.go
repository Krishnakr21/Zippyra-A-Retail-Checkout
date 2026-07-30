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
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/middleware"

	"github.com/gorilla/mux"
)

const (
	RoleCustomer     = "CUSTOMER"
	RoleCashier      = "CASHIER"
	RoleStoreManager = "STORE_MANAGER"
	RoleAdmin        = "ADMIN"
	RoleSystem       = "SYSTEM"
)

func main() {
	cfg, err := config.Load("order-service")
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

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers)

	// Initialize Services & Repositories
	repo := NewPostgresRepository(database.DB)
	exitTokenSvc := NewMockRedisExitTokenService(cfg.JWT.Secret)
	invoiceSvc := NewRealInvoiceService(repo)
	_ = NewGSTIRNConsumer(repo, invoiceSvc)

	// Initialize Outbox Relay Worker
	outboxRelay := NewOutboxRelay(database.DB, producer, 200*time.Millisecond)
	outboxRelay.Start()
	defer outboxRelay.Stop()

	// Handlers & Router Setup
	handler := NewOrderHandler(repo, exitTokenSvc, invoiceSvc, cfg.JWT.Secret)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RecoverMiddleware)

	RegisterRoutes(r, handler, outboxRelay)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Order Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Order Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}
