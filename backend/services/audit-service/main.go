package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
)

func main() {
	logger.Info("Starting audit-service...")

	database, err := db.ConnectDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := NewPostgresRepository(database)
	kafkaAdmin := NewKafkaAdminClient(os.Getenv("KAFKA_BROKERS"))
	jwtSecret := os.Getenv("JWT_SECRET")
	handler := NewAuditHandler(repo, kafkaAdmin, jwtSecret, nil)
	consumer := NewAuditConsumer(repo)
	_ = consumer // Consumer would be started via Kafka consumer group in production

	router := SetupRoutes(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8015"
	}

	logger.Info("audit-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
