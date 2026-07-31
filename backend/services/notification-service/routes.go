package main

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
)

func SetupRoutes(
	r *mux.Router,
	handler *NotificationHandler,
	adminHandler *AdminHandler,
	db *sql.DB,
) {
	// Customer / Staff Device Tokens & Preferences
	r.HandleFunc("/v1/notification/device-tokens", handler.RegisterDeviceToken).Methods("POST")
	r.HandleFunc("/v1/notification/device-tokens/{device_id}", handler.DeactivateDeviceToken).Methods("DELETE")

	r.HandleFunc("/v1/notification/preferences", handler.GetPreferences).Methods("GET")
	r.HandleFunc("/v1/notification/preferences", handler.UpdatePreference).Methods("PUT")

	r.HandleFunc("/v1/notification/inbox", handler.GetInbox).Methods("GET")
	r.HandleFunc("/v1/notification/inbox/unread-count", handler.GetUnreadCount).Methods("GET")
	r.HandleFunc("/v1/notification/inbox/{id}/read", handler.MarkRead).Methods("PUT")

	// Admin Endpoints
	r.HandleFunc("/v1/notification/admin/whatsapp-templates", adminHandler.ListWhatsAppTemplates).Methods("GET")
	r.HandleFunc("/v1/notification/admin/whatsapp-templates/{key}", adminHandler.UpdateWhatsAppTemplate).Methods("PUT")

	r.HandleFunc("/v1/notification/admin/ops-alert-channels", adminHandler.ListOpsAlertChannels).Methods("GET")
	r.HandleFunc("/v1/notification/admin/ops-alert-channels", adminHandler.CreateOpsAlertChannel).Methods("POST")
	r.HandleFunc("/v1/notification/admin/ops-alert-channels/{id}", adminHandler.UpdateOpsAlertChannel).Methods("PUT")

	// Health check probes
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")
}
