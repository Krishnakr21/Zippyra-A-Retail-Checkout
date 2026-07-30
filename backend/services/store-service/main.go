package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/redis"
)

func main() {
	// 1. Load & validate configuration
	cfg, err := config.Load("store-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	cfg.LogActiveIntegrations()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Hot reload secrets in staging/production
	config.StartSecretsManagerHotReload(ctx, cfg.Environment, cfg.ServiceName, cfg.AWSRegion, cfg.AWS.SecretsRotationPollMinutes)

	// 2. Connect DB
	database, dbErr := db.ConnectDB(cfg.Database.URL)
	var repo Repository
	if dbErr == nil && database != nil && database.DB != nil {
		repo = NewPostgresRepository(database)
		logger.Info("Connected to PostgreSQL database")
	} else {
		logger.Info("Using in-memory repository fallback for store-service")
		repo = NewMemoryRepository()
	}

	// 3. Connect Redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Realtime.Host, cfg.Redis.Realtime.Port)
	rdb, _ := redis.ConnectRedis(redisAddr)

	var capacityMgr CapacityManager
	if rdb != nil && rdb.Client != nil {
		capacityMgr = NewRedisCapacityManager(rdb.Client)
		logger.Info("Connected to Redis realtime cluster at %s", redisAddr)
	} else {
		capacityMgr = NewMemoryCapacityManager()
		logger.Info("Using in-memory capacity manager fallback")
	}

	// 4. Session Manager
	sessionMgr := NewSessionManager(repo, capacityMgr, cfg.JWT.Secret, nil)

	// 5. Start Background Auto-Expire Worker (ticks every 5 min)
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				sessionMgr.AutoExpireStaleSessionsJob(ctx)
			}
		}
	}()

	// 6. Handlers & Router
	customerHandler := NewStoreHandler(repo, capacityMgr, sessionMgr, cfg.JWT.Secret)
	internalHandler := NewInternalAdminWriteHandler(repo, cfg.JWT.Secret)
	selfManageHandler := NewSelfManageHandler(repo, cfg.JWT.Secret)
	router := SetupRoutes(customerHandler, internalHandler, selfManageHandler)

	// 7. HTTP Server & Graceful Shutdown
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		logger.Info("store-service running on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	// Graceful shutdown with 25s budget
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down store-service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown: %v", err)
	} else {
		logger.Info("store-service stopped cleanly")
	}
}
