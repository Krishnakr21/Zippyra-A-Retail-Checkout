package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type SyncJobHandler struct {
	repo             IntegrationRepository
	directPushWorker *DirectPushWorker
	jwtSecret        string
}

func NewSyncJobHandler(repo IntegrationRepository, directPushWorker *DirectPushWorker) *SyncJobHandler {
	return &SyncJobHandler{
		repo:             repo,
		directPushWorker: directPushWorker,
		jwtSecret:        "dev-secret-key-change-in-prod",
	}
}

func (h *SyncJobHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil {
			return claims, nil
		}
	}

	role := r.Header.Get("X-User-Role")
	chainID := r.Header.Get("X-Chain-ID")
	userID := r.Header.Get("X-User-ID")
	if role != "" {
		return &jwt.Claims{
			UserID:  userID,
			Role:    role,
			ChainID: chainID,
		}, nil
	}
	return nil, sharedErrors.NewAPIError(sharedErrors.CodeUnauthorized, "Unauthorized", nil)
}

// 12. GET /v1/integration/connections/{id}/sync-jobs
func (h *SyncJobHandler) ListSyncJobs(w http.ResponseWriter, r *http.Request) {
	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	connID := vars["id"]

	var statusFilter *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = &s
	}
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "OUTBOUND"
	}

	jobs, err := h.repo.ListSyncJobs(r.Context(), connID, statusFilter, direction)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list sync jobs", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"jobs": jobs})
}

// 13. GET /v1/integration/connections/{id}/webhook-events
func (h *SyncJobHandler) ListWebhookEvents(w http.ResponseWriter, r *http.Request) {
	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	connID := vars["id"]

	var resultFilter *string
	if res := r.URL.Query().Get("result"); res != "" {
		resultFilter = &res
	}

	events, err := h.repo.ListWebhookEvents(r.Context(), connID, resultFilter)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list webhook events", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"events": events})
}

// 14. POST /v1/integration/connections/{id}/sync-jobs/{job_id}/retry
func (h *SyncJobHandler) RetrySyncJob(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	connID := vars["id"]
	jobID := vars["job_id"]

	if claims.Role != "OWNER" && claims.Role != "SUPER_ADMIN" && claims.Role != "OPERATIONS" {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Retry sync job requires operational privileges", nil)
		return
	}

	conn, err := h.repo.GetConnectionByID(r.Context(), connID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	job, err := h.repo.GetSyncJobByID(r.Context(), jobID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Sync job not found", nil)
		return
	}

	if conn.IntegrationMode != IntegrationModeDirect {
		sharedErrors.WriteError(w, http.StatusBadRequest, "DIRECT_MODE_ONLY", "Manual sync job retry is only applicable for DIRECT connections", nil)
		return
	}

	pushErr := h.directPushWorker.PushSyncJob(r.Context(), job, conn)
	if pushErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "RETRY_FAILED",
			"failure_reason": pushErr.Error(),
			"attempt_count":  job.AttemptCount,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "RETRY_SUCCESS",
		"attempt_count": job.AttemptCount,
	})
}
