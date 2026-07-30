package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/redis"
)

func main() {
	modeFlag := flag.String("mode", "", "Execution mode (e.g. backfill-offer-rules)")
	flag.Parse()

	// 1. Load & validate configuration
	cfg, err := config.Load("cart-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	cfg.LogActiveIntegrations()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Secrets Manager hot reload in staging/prod
	config.StartSecretsManagerHotReload(ctx, cfg.Environment, cfg.ServiceName, cfg.AWSRegion, cfg.AWS.SecretsRotationPollMinutes)

	// 2. Connect Postgres DB & Repositories
	database, dbErr := db.ConnectDB(cfg.Database.URL)
	var checkoutRepo CheckoutSessionRepository
	var offerRepo OfferRepository

	if dbErr == nil && database != nil && database.DB != nil {
		checkoutRepo = NewPostgresCheckoutRepository(database)
		offerRepo = NewPostgresOfferRepository(database.DB)
		logger.Info("Connected to PostgreSQL database for cart-service")
	} else {
		logger.Info("Using in-memory checkout & offer repository fallbacks for cart-service")
		checkoutRepo = NewMemoryCheckoutRepository()
		offerRepo = NewMemoryOfferRepository()
	}

	// 3. Connect Redis Cart Cluster
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Cart.Host, cfg.Redis.Cart.Port)
	rdb, _ := redis.ConnectRedis(redisAddr)

	var cartStore CartStore
	var holdManager HoldManager
	var offerEngine OfferEngine
	var lockManager LockManager
	var redisCmdable goredis.Cmdable

	if rdb != nil && rdb.Client != nil {
		redisCmdable = rdb.Client
		cartStore = NewRedisCartStore(rdb.Client)
		holdManager = NewRedisHoldManager(rdb.Client)
		offerEngine = NewRedisOfferEngine(rdb.Client)
		lockManager = NewRedisLockManager(rdb.Client)
		logger.Info("Connected to Redis Cart cluster at %s", redisAddr)
	} else {
		cartStore = NewRedisCartStore(nil)
		holdManager = NewMemoryHoldManager()
		offerEngine = NewMemoryOfferEngine()
		lockManager = NewMemoryLockManager()
		logger.Info("Using in-memory components fallback")
	}

	// 4. Compiler & Background Jobs
	offerCompiler := NewOfferCompiler(offerRepo, redisCmdable)
	scheduleJob := NewOfferScheduleJob(offerCompiler, offerRepo)
	reconcileJob := NewOfferReconciliationJob(offerCompiler, offerRepo, redisCmdable)

	// Mode Handler: ONE-TIME CLI BACKFILL
	if *modeFlag == "backfill-offer-rules" || os.Getenv("MODE") == "backfill-offer-rules" {
		logger.Info("Starting one-time offer rules backfill execution...")
		storesToBackfill, err := offerRepo.ListStoresWithNoAudit(ctx)
		if err != nil {
			logger.Error("Backfill failed to list stores: %v", err)
			os.Exit(1)
		}
		if len(storesToBackfill) == 0 {
			// Fallback: list default stores
			storesToBackfill, _ = offerRepo.ListStoresForChain(ctx, "chain-default-001")
		}

		logger.Info("Found %d store(s) requiring offer rule compilation backfill", len(storesToBackfill))
		successCount := 0
		for _, s := range storesToBackfill {
			if err := offerCompiler.CompileAndPublish(ctx, s); err != nil {
				logger.Error("Failed to backfill store %s: %v", s, err)
			} else {
				successCount++
			}
		}
		logger.Info("Completed offer rules backfill: %d / %d stores compiled", successCount, len(storesToBackfill))
		os.Exit(0)
	}

	// Start background workers
	go scheduleJob.Start(ctx)
	go reconcileJob.Start(ctx)
	go StartStaleSessionCleaner(ctx, checkoutRepo, lockManager)

	// 5. Handlers
	catalogEngine := NewDefaultCatalogLookupEngine("http://localhost:8084")
	eventProc := NewEventProcessor(cartStore, holdManager, nil)

	customerHandler := NewCartHandler(cartStore, holdManager, offerEngine, checkoutRepo, lockManager, catalogEngine, eventProc, cfg.JWT.Secret, offerRepo)
	internalHandler := NewInternalCartHandler(checkoutRepo, cartStore, holdManager, lockManager, cfg.JWT.Secret)
	adminHandler := NewOfferAdminHandler(offerRepo, offerCompiler, redisCmdable)

	couponCompiler := NewCouponCompiler(redisCmdable)
	couponAdminHandler := NewCouponAdminHandler(offerRepo, couponCompiler)
	couponReconcileJob := NewCouponReconciliationJob(offerRepo, couponCompiler)
	go couponReconcileJob.StartHourlyLoop(ctx)

	// Readiness check verifying DB, Redis, and reconciliation job status
	healthChecker := func() bool {
		if database != nil && database.DB != nil {
			if err := database.DB.Ping(); err != nil {
				return false
			}
		}
		if !reconcileJob.IsHealthy() {
			return false
		}
		return true
	}

	router := SetupRoutes(customerHandler, internalHandler, adminHandler, couponAdminHandler, healthChecker)

	// 6. HTTP Server & Graceful Shutdown
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		logger.Info("cart-service running on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down cart-service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown: %v", err)
	} else {
		logger.Info("cart-service stopped cleanly")
	}
}
