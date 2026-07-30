package main

import (
	"context"
	"fmt"
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
	"github.com/zippyra/backend/shared/redis"
)

const (
	RoleCustomer     = "CUSTOMER"
	RoleCashier      = "CASHIER"
	RoleSecurity     = "SECURITY"
	RoleStoreManager = "STORE_MANAGER"
	RoleAdmin        = "ADMIN"
	RoleSystem       = "SYSTEM"
)

func main() {
	cfg, err := config.Load("exit-validation-service")
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

	// Initialize Redis Client (ExitToken cluster)
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.ExitToken.Host, cfg.Redis.ExitToken.Port)
	redisClient, err := redis.ConnectRedis(redisAddr)
	if err != nil {
		logger.Error("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers)

	// Initialize MQTT Client & Verifiers
	mqttClient := NewMockMQTTClient()
	repo := NewPostgresRepository(database.DB)
	verifier := NewJWTVerifier(cfg.JWT.Secret, NewRedisRevocationChecker(redisClient))
	metrics := NewAlarmMetrics()

	handler := NewExitHandler(repo, redisClient, producer, mqttClient, verifier, metrics, cfg.JWT.Secret)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RecoverMiddleware)

	RegisterRoutes(r, handler, database.DB, redisClient, mqttClient)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Exit Validation Service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Exit Validation Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
}
