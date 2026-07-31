package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/errors"
)

type AuditHandler struct {
	repo           Repository
	kafkaAdmin     *KafkaAdminClient
	jwtSecret      string
	auditPublisher *audit.Publisher
}

func NewAuditHandler(repo Repository, kafkaAdmin *KafkaAdminClient, jwtSecret string, auditPub *audit.Publisher) *AuditHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	if kafkaAdmin == nil {
		kafkaAdmin = NewKafkaAdminClient("localhost:9092")
	}
	return &AuditHandler{
		repo:           repo,
		kafkaAdmin:     kafkaAdmin,
		jwtSecret:      jwtSecret,
		auditPublisher: auditPub,
	}
}

func (h *AuditHandler) HandleListActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	filter := AuditFilter{
		ActorID:    r.URL.Query().Get("actor_id"),
		ActionType: r.URL.Query().Get("action_type"),
		TargetType: r.URL.Query().Get("target_type"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
		Page:       page,
		PageSize:   pageSize,
	}

	actions, total, err := h.repo.ListActions(r.Context(), filter)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list audit actions", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"actions":   actions,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}
