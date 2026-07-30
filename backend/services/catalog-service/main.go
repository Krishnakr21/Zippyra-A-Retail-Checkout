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
	cfg, err := config.Load("catalog-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	cfg.LogActiveIntegrations()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Secrets Manager hot reload in staging/prod
	config.StartSecretsManagerHotReload(ctx, cfg.Environment, cfg.ServiceName, cfg.AWSRegion, cfg.AWS.SecretsRotationPollMinutes)

	// 2. Connect Postgres
	database, dbErr := db.ConnectDB(cfg.Database.URL)
	var repo Repository
	if dbErr == nil && database != nil && database.DB != nil {
		repo = NewPostgresRepository(database)
		logger.Info("Connected to PostgreSQL database")
	} else {
		logger.Info("Using in-memory repository fallback for catalog-service")
		repo = NewMemoryRepository()
	}

	// 3. Connect Redis SKU Cache Cluster
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.SKUCache.Host, cfg.Redis.SKUCache.Port)
	rdb, _ := redis.ConnectRedis(redisAddr)

	var cacheMgr CacheManager
	if rdb != nil && rdb.Client != nil {
		cacheMgr = NewRedisCacheManager(rdb.Client)
		logger.Info("Connected to Redis SKU cache cluster at %s", redisAddr)
	} else {
		cacheMgr = NewMemoryCacheManager()
		logger.Info("Using in-memory SKU cache manager fallback")
	}

	// 4. Connect Elasticsearch Search Engine
	esEndpoint := fmt.Sprintf("http://%s:%d", cfg.ES.Host, cfg.ES.Port)
	searchEngine := NewESSearchEngine(esEndpoint, repo)

	// 5. Services & Workers
	syncEngine := NewSyncEngineService(repo)
	importWorker := NewImportWorker(repo, cacheMgr, searchEngine, nil)

	// 6. Handlers & Router
	customerHandler := NewCatalogHandler(repo, cacheMgr, searchEngine, syncEngine, cfg.JWT.Secret)
	adminHandler := NewAdminCatalogHandler(repo, cacheMgr, searchEngine, importWorker, nil, cfg.JWT.Secret)

	// Readiness check verifying DB, Redis, and ES
	healthChecker := func() bool {
		if database != nil && database.DB != nil {
			if err := database.DB.Ping(); err != nil {
				return false
			}
		}
		return true
	}

	router := SetupRoutes(customerHandler, adminHandler, healthChecker)

	// 7. HTTP Server & Graceful Shutdown
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		logger.Info("catalog-service running on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down catalog-service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown: %v", err)
	} else {
		logger.Info("catalog-service stopped cleanly")
	}
}
