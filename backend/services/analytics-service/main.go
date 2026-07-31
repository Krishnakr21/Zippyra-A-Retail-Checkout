package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zippyra/backend/shared/logger"
)

func main() {
	logger.Info("Starting analytics-service...")

	repo := NewMemoryRepository()
	dedupGuard := NewMemoryHourlyDedupGuard()

	_ = NewAnalyticsConsumer(repo, dedupGuard)

	h := NewAnalyticsHandler(repo)
	router := SetupRoutes(h)

	port := os.Getenv("ANALYTICS_SERVICE_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8092"
	}

	logger.Info("analytics-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
