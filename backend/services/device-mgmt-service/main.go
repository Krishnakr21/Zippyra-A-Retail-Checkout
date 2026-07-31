package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	logger.Info("Starting device-mgmt-service...")

	var repo Repository
	database, err := db.ConnectDB(os.Getenv("DATABASE_URL"))
	if err != nil || database == nil {
		logger.Warn("Postgres/TimescaleDB unavailable, using MemoryRepository fallback")
		repo = NewMemoryRepository()
	} else {
		defer database.Close()
		repo = NewPostgresRepository(database)
	}

	iotProvider := NewMockIoTProvider()
	auditPub := audit.NewPublisher(nil, "device-mgmt-service")

	deviceH := NewDeviceHandler(repo, iotProvider, nil, auditPub)
	router := SetupRoutes(deviceH)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8017"
	}

	logger.Info("device-mgmt-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
