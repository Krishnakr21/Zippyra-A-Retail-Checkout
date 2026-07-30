package main

import (
	"net/http"
	"strings"

	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(handler *AuditHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz/live", health.LiveHandler)
	mux.HandleFunc("/healthz/ready", health.ReadyHandler)
	mux.HandleFunc("/healthz/startup", health.StartupHandler)

	mux.HandleFunc("/v1/audit/actions", handler.HandleListActions)

	// Kafka DLQ Routes
	mux.HandleFunc("/v1/audit/kafka/dlq-topics", handler.HandleListDLQTopics)
	mux.HandleFunc("/v1/audit/kafka/dlq-topics/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/replay") {
			handler.HandleReplayDLQMessages(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/messages") {
			if r.Method == http.MethodDelete {
				handler.HandleDiscardDLQMessages(w, r)
			} else {
				handler.HandlePeekDLQMessages(w, r)
			}
		} else {
			handler.HandlePeekDLQMessages(w, r)
		}
	})

	// Feature Flag Admin Routes
	mux.HandleFunc("/v1/audit/feature-flags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.HandleCreateFeatureFlag(w, r)
		} else {
			handler.HandleListFeatureFlags(w, r)
		}
	})
	mux.HandleFunc("/v1/audit/feature-flags/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			handler.HandleUpdateFeatureFlag(w, r)
		} else if r.Method == http.MethodDelete {
			handler.HandleDeleteFeatureFlag(w, r)
		} else {
			handler.HandleListFeatureFlags(w, r)
		}
	})

	h := middleware.MaxBytesMiddleware(1048576)(mux)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
