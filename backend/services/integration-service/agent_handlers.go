package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	sharedErrors "github.com/zippyra/backend/shared/errors"
)

type AgentHandler struct {
	repo IntegrationRepository
}

func NewAgentHandler(repo IntegrationRepository) *AgentHandler {
	return &AgentHandler{repo: repo}
}

func (h *AgentHandler) authenticateAgent(r *http.Request, conn *ERPConnection) error {
	if conn.IntegrationMode == IntegrationModeDirect {
		return errors.New("DIRECT_MODE_ONLY")
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("AGENT_AUTH_FAILED")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	tokenHash := hashString(token)

	if conn.AgentAPIKeyHash == nil || *conn.AgentAPIKeyHash != tokenHash {
		return errors.New("AGENT_AUTH_FAILED")
	}

	return nil
}

// 10. GET /v1/integration/connections/{id}/pull-queue?limit=50
func (h *AgentHandler) PullQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connID := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), connID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	if authErr := h.authenticateAgent(r, conn); authErr != nil {
		if authErr.Error() == "DIRECT_MODE_ONLY" {
			sharedErrors.WriteError(w, http.StatusBadRequest, "DIRECT_MODE_ONLY", "Agent pull-queue is for AGENT_POLLED connections only", nil)
			return
		}
		sharedErrors.WriteError(w, http.StatusUnauthorized, "AGENT_AUTH_FAILED", "Invalid agent API key", nil)
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	jobs, err := h.repo.ListPendingSyncJobs(r.Context(), connID, limit)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to fetch pending sync jobs", nil)
		return
	}

	// Mark returned jobs as DELIVERED
	if len(jobs) > 0 {
		var jobIDs []string
		for _, j := range jobs {
			jobIDs = append(jobIDs, j.ID)
		}
		_ = h.repo.MarkSyncJobsDelivered(r.Context(), jobIDs)
	}

	// Update connection timestamps
	now := time.Now()
	var newStatus *ConnectionStatus
	if conn.Status == ConnectionStatusPendingSetup {
		active := ConnectionStatusActive
		newStatus = &active
	}
	_ = h.repo.UpdateConnectionTimestamps(r.Context(), connID, nil, nil, &now, newStatus)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// 11. POST /v1/integration/connections/{id}/pull-queue/ack
func (h *AgentHandler) AckQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connID := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), connID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	if authErr := h.authenticateAgent(r, conn); authErr != nil {
		if authErr.Error() == "DIRECT_MODE_ONLY" {
			sharedErrors.WriteError(w, http.StatusBadRequest, "DIRECT_MODE_ONLY", "Agent pull-queue is for AGENT_POLLED connections only", nil)
			return
		}
		sharedErrors.WriteError(w, http.StatusUnauthorized, "AGENT_AUTH_FAILED", "Invalid agent API key", nil)
		return
	}

	var req AckPullQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid request payload", nil)
		return
	}

	if len(req.JobIDs) > 0 {
		_ = h.repo.MarkSyncJobsAcknowledged(r.Context(), connID, req.JobIDs)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"acknowledged_count": len(req.JobIDs),
	})
}

func parseHeaderToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

var _ = fmt.Sprintf
