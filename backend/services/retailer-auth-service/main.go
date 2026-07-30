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
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/middleware"
	"github.com/zippyra/backend/shared/redis"
	"github.com/zippyra/backend/shared/sms"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}

	dbConn, err := db.ConnectDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Database connection failed: %v", err)
	}
	defer dbConn.Close()

	redisClient, err := redis.ConnectRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Warn("Redis connection failed: %v", err)
	}

	repo := NewPostgresRepository(dbConn)
	smsSender := sms.NewTwilioSmsSender()
	otpSvc := NewOTPService(repo, redisClient, smsSender)
	pinSvc := NewPINService(repo)
	shiftSvc := NewShiftService(repo)

	staffH := NewStaffHandler(repo, jwtSecret)
	authH := NewAuthHandler(repo, otpSvc, pinSvc, jwtSecret)
	shiftH := NewShiftHandler(shiftSvc, jwtSecret)

	router := mux.NewRouter()
	RegisterRoutes(router, staffH, authH, shiftH, func() bool {
		return dbConn.PingContext(context.Background()) == nil
	})

	handler := middleware.MaxBytesMiddleware(1048576)(router)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("Starting retailer-auth-service on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down retailer-auth-service gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}
}
