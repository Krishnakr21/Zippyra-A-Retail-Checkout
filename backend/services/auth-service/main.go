package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/redis"
)

func main() {
	// 1. Load & validate configuration first thing (fail-fast)
	cfg, err := config.Load("auth-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	// Log active vs stubbed integrations
	cfg.LogActiveIntegrations()

	// Hot reload secrets in staging/production
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config.StartSecretsManagerHotReload(ctx, cfg.Environment, cfg.ServiceName, cfg.AWSRegion, cfg.AWS.SecretsRotationPollMinutes)

	// 2. Connect DB
	database, dbErr := db.ConnectDB(cfg.Database.URL)
	var repo Repository
	if dbErr == nil && database != nil && database.DB != nil {
		repo = NewPostgresRepository(database)
		logger.Info("Connected to PostgreSQL database")
	} else {
		logger.Info("Using in-memory repository fallback for auth-service")
		repo = NewMemoryRepository()
	}

	// 3. Connect Redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.RateLimit.Host, cfg.Redis.RateLimit.Port)
	rdb, _ := redis.ConnectRedis(redisAddr)

	// 4. Instantiate Senders & Managers
	var smsSender SmsSender
	if cfg.OTP.SMSProvider == "twilio" {
		smsSender = NewTwilioSmsSender()
	} else {
		smsSender = &LogSmsSender{}
	}

	emailSender := NewGmailEmailSender()
	otpMgr := NewDefaultOTPManager(rdb, smsSender, emailSender)
	googleVerifier := &RealGoogleTokenVerifier{}

	// 5. Setup Handlers & Routes
	handler := NewAuthHandler(repo, otpMgr, googleVerifier, rdb)
	router := SetupRoutes(handler)

	// 6. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("auth-service running on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Error("Server stopped: %v", err)
	}
}
