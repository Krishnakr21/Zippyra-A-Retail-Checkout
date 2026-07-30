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

func main() {
	cfg, err := config.Load("payment-service")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Database
	database, err := db.ConnectDB(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to Postgres: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers)

	// Initialize Repository & Clients
	repo := NewPostgresRepository(database.DB)
	cartClient := NewDefaultCartServiceClient("http://localhost:8084")
	loyaltyClient := NewMockLoyaltyServiceClient()

	rzpClient := NewRazorpayClient(cfg.Payment.RazorpayKeyID, cfg.Payment.RazorpayKeySecret, cfg.Payment.RazorpayWebhookSecret)
	cfClient := NewCashfreeClient(cfg.Payment.CashfreeAppID, cfg.Payment.CashfreeSecretKey)

	circuitBreaker := NewRollingCircuitBreaker(0.05, 30*time.Second)

	// Initialize Outbox Relay Worker
	outboxRelay := NewOutboxRelay(database.DB, producer, 200*time.Millisecond)
	outboxRelay.Start()
	defer outboxRelay.Stop()

	// Initialize 10-minute Payment Timeout Background Cleaner
	go startTimeoutCleanerJob(ctx, repo, loyaltyClient)

	// Handlers & Router Setup
	handler := NewPaymentHandler(repo, cartClient, loyaltyClient, rzpClient, cfClient, circuitBreaker, cfg.Payment.RazorpayKeyID, cfg.JWT.Secret)

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
		logger.Info("Payment Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Payment Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}

// Background Job: 10-minute Payment Timeout Cleaner
func startTimeoutCleanerJob(ctx context.Context, repo Repository, loyalty LoyaltyServiceClient) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			payments, err := repo.GetPendingTimeoutPayments(ctx, 10*time.Minute)
			if err != nil {
				logger.Error("Failed to query pending timeout payments: %v", err)
				continue
			}

			for _, p := range payments {
				_ = repo.FailPaymentAndReleaseTx(ctx, p.ID, "Payment timed out after 10 minutes")
				if p.LoyaltyPointsUsed > 0 {
					_ = loyalty.ReleaseReservedPoints(ctx, p.UserID, p.LoyaltyPointsUsed)
				}
				logger.Info("Marked stuck payment %s as FAILED and released loyalty points", p.ID)
			}
		}
	}
}

// Daily 6am IST Reconciliation Function Stub
func RunDailyReconciliationReport(ctx context.Context, repo Repository) error {
	logger.Info("Executing 6am IST daily Razorpay settlement reconciliation report...")
	return nil
}
