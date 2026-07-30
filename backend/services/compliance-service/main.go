package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	cfg := LoadConfig()

	// Initialize Database
	database, err := db.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.KafkaBrokers)

	// Initialize GSP IRP Client
	irpClient := NewHTTPIRPClient(cfg.GSTIRPBaseURL, cfg.GSTIRPUsername, cfg.GSTIRPPassword, cfg.GSTIRPClientID)

	// Initialize Repository & Services
	repo := NewMemoryRepository() // Falls back to Postgres when DB is active
	consentSvc := NewDPDPConsentService(repo)
	requestSvc := NewDPDPRequestService(repo)
	deletionProc := NewDPDPDeletionProcessor(repo, producer)
	accessMgr := NewAccessExportManager(repo, nil)
	accessHandler := NewAccessFulfillmentHandler(accessMgr)
	kycSvc := NewKYCService(repo)
	reconJob := NewReconciliationJob(repo, cfg.PaymentServiceURL)
	irnRetryJob := NewIRNRetryJob(repo, irpClient)

	// Start Background Retry Worker (5 minute interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go irnRetryJob.StartWorker(ctx, 5*time.Minute)

	// Start Background Export Expiry Sweeper (1 hour interval)
	exportExpiryJob := NewExportExpiryJob(repo, 1*time.Hour)
	exportExpiryJob.Start()
	defer exportExpiryJob.Stop()

	handler := NewComplianceHandler(repo, consentSvc, requestSvc, deletionProc, accessHandler, kycSvc, reconJob, irnRetryJob, cfg.JWTSecret)

	router := mux.NewRouter()
	RegisterRoutes(router, handler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("compliance-service starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down compliance-service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}
