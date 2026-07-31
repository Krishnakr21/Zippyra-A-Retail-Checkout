package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/audit"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type ReplayRequest struct {
	Offsets []int64 `json:"offsets"`
}

type DiscardRequest struct {
	Offsets []int64 `json:"offsets"`
	Reason  string  `json:"reason"`
}

func (h *AuditHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing authorization header")
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	return jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
}

// GET /v1/audit/kafka/dlq-topics
func (h *AuditHandler) HandleListDLQTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	topics, err := h.kafkaAdmin.ListDLQTopics(r.Context())
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list DLQ topics", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"dlq_topics": topics,
	})
}

// GET /v1/audit/kafka/dlq-topics/{topic}/messages?limit={n}
func (h *AuditHandler) HandlePeekDLQMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid topic path", nil)
		return
	}
	topic := pathParts[5]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	discardedMap, err := h.repo.GetDiscardedOffsets(r.Context(), topic)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to fetch discarded offsets", nil)
		return
	}

	msgs, err := h.kafkaAdmin.PeekDLQMessages(r.Context(), topic, limit, discardedMap)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to peek DLQ messages", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"topic":    topic,
		"messages": msgs,
		"total":    len(msgs),
	})
}

// POST /v1/audit/kafka/dlq-topics/{topic}/replay
func (h *AuditHandler) HandleReplayDLQMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid topic path", nil)
		return
	}
	topic := pathParts[5]

	var req ReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Offsets) == 0 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid offsets payload", nil)
		return
	}

	replayedCount, failedOffsets, err := h.kafkaAdmin.ReplayDLQMessages(r.Context(), topic, req.Offsets)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to replay DLQ messages", nil)
		return
	}

	// Audit Event Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "audit.dlq_message_replayed",
			TargetType: "dlq_topic",
			TargetID:   topic,
			Payload: map[string]interface{}{
				"replayed_count": replayedCount,
				"failed_offsets": failedOffsets,
				"offsets":        req.Offsets,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"replayed_count": replayedCount,
		"failed_offsets": failedOffsets,
	})
}

// DELETE /v1/audit/kafka/dlq-topics/{topic}/messages (Soft discard)
func (h *AuditHandler) HandleDiscardDLQMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid topic path", nil)
		return
	}
	topic := pathParts[5]

	var req DiscardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Offsets) == 0 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid offsets payload", nil)
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	err = h.repo.DiscardOffsets(r.Context(), topic, req.Offsets, userID, req.Reason)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to discard offsets", nil)
		return
	}

	// Audit Event Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "audit.dlq_message_discarded",
			TargetType: "dlq_topic",
			TargetID:   topic,
			Payload: map[string]interface{}{
				"discarded_offsets": req.Offsets,
				"reason":            req.Reason,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "DISCARDED",
		"discarded_offsets": req.Offsets,
	})
}
