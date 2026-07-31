package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/config"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/middleware"
	"github.com/zippyra/backend/shared/redis"
)

func main() {
	modeFlag := flag.String("mode", "server", "Execution mode: 'server' or 'shrinkage-rollup'")
	dateFlag := flag.String("date", "", "Target date for shrinkage rollup (YYYY-MM-DD), default yesterday")
	flag.Parse()

	cfg, err := config.Load("inventory-service")
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

	producer := kafka.NewProducer(cfg.Kafka.Brokers)
	repo := NewPostgresRepository(database.DB)
	engine := NewMovementEngine(database.DB, producer, repo.IsSQLite())

	// CLI Mode Check: Shrinkage Rollup Job
	if *modeFlag == "shrinkage-rollup" {
		job := NewShrinkageRollupJob(database.DB, repo, producer)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := job.Run(ctx, *dateFlag); err != nil {
			logger.Error("Shrinkage rollup job failed: %v", err)
			os.Exit(1)
		}
		logger.Info("Shrinkage rollup job completed successfully")
		os.Exit(0)
	}

	// Server Mode
	redisAddr := cfg.Redis.SKUCache.Host + ":" + strconv.Itoa(cfg.Redis.SKUCache.Port)
	redisClient, _ := redis.ConnectRedis(redisAddr)
	if redisClient != nil {
		defer redisClient.Close()
	}

	consumer := NewEventConsumer(engine)
	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, "inventory-service-group", []string{"order.completed", "order.returned"})
	kafkaConsumer.RegisterHandler("order.completed", consumer.ProcessOrderCompleted)
	kafkaConsumer.RegisterHandler("order.returned", consumer.ProcessOrderReturned)

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go kafkaConsumer.Start(consumerCtx)

	handler := NewInventoryHandler(repo, engine, redisClient)
	internalHandler := NewInternalHandler(repo, engine)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RecoverMiddleware)

	RegisterRoutes(r, handler, internalHandler, database.DB, cfg.JWT.Secret)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Inventory Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Inventory Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}
