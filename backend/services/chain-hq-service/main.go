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
	logger.Info("Starting chain-hq-service...")

	var repo Repository
	database, err := db.ConnectDB(os.Getenv("DATABASE_URL"))
	if err != nil || database == nil {
		logger.Warn("Database unavailable, using MemoryRepository fallback")
		repo = NewMemoryRepository()
	} else {
		defer database.Close()
		repo = NewPostgresRepository(database)
	}

	auditPub := audit.NewPublisher(nil, "chain-hq-service")

	authH := NewAuthHandler(repo)
	userH := NewUserManagementHandler(repo, auditPub)
	dashH := NewDashboardHandler()
	bulkH := NewBulkImportHandler(repo, auditPub)

	router := SetupRoutes(authH, userH, dashH, bulkH)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8016"
	}

	logger.Info("chain-hq-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
