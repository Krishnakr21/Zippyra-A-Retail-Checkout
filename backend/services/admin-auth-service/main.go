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
	logger.Info("Starting admin-auth-service...")

	var repo Repository
	database, err := db.ConnectDB(os.Getenv("DATABASE_URL"))
	if err != nil || database == nil {
		logger.Warn("Database unavailable, starting with MemoryRepository fallback")
		repo = NewMemoryRepository()
	} else {
		defer database.Close()
		repo = NewPostgresRepository(database)
	}

	googleVal := &RealGoogleTokenValidator{}
	auditPub := audit.NewPublisher(nil, "admin-auth-service")

	authHandler := NewAdminAuthHandler(repo, googleVal, auditPub)
	mgmtHandler := NewAdminManagementHandler(repo)

	router := SetupRoutes(authHandler, mgmtHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8013"
	}

	logger.Info("admin-auth-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
