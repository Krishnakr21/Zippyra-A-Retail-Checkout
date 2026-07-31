package main

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
)

func SetupRoutes(
	r *mux.Router,
	connHandler *ConnectionHandler,
	webhookHandler *WebhookHandler,
	agentHandler *AgentHandler,
	syncJobHandler *SyncJobHandler,
	db *sql.DB,
) {
	// Connection Management
	r.HandleFunc("/v1/integration/connections", connHandler.CreateConnection).Methods("POST")
	r.HandleFunc("/v1/integration/connections", connHandler.ListConnections).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}", connHandler.GetConnection).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}", connHandler.UpdateConnection).Methods("PUT")
	r.HandleFunc("/v1/integration/connections/{id}", connHandler.DeleteConnection).Methods("DELETE")
	r.HandleFunc("/v1/integration/connections/{id}/rotate-secret", connHandler.RotateSecret).Methods("POST")

	// Inbound Webhook
	r.HandleFunc("/v1/integration/connections/{id}/webhook", webhookHandler.HandleInboundWebhook).Methods("POST")

	// Agent Polling APIs
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue", agentHandler.PullQueue).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue/ack", agentHandler.AckQueue).Methods("POST")

	// Sync Jobs & Webhook Events Logs
	r.HandleFunc("/v1/integration/connections/{id}/sync-jobs", syncJobHandler.ListSyncJobs).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}/webhook-events", syncJobHandler.ListWebhookEvents).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}/sync-jobs/{job_id}/retry", syncJobHandler.RetrySyncJob).Methods("POST")

	// Health check probes
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")
}
