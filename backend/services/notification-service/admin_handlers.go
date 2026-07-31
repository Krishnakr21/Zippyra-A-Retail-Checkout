package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type AdminHandler struct {
	repo      NotificationRepository
	jwtSecret string
}

func NewAdminHandler(repo NotificationRepository) *AdminHandler {
	return &AdminHandler{
		repo:      repo,
		jwtSecret: "dev-secret-key-change-in-prod",
	}
}

func (h *AdminHandler) extractAdminClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil {
			return claims, nil
		}
	}

	role := r.Header.Get("X-User-Role")
	if role == "SUPER_ADMIN" || role == "ADMIN" || role == "OPERATIONS" {
		return &jwt.Claims{Role: role}, nil
	}
	return nil, sharedErrors.NewAPIError(sharedErrors.CodeUnauthorized, "Unauthorized", nil)
}

// 7. GET /v1/notification/admin/whatsapp-templates
func (h *AdminHandler) ListWhatsAppTemplates(w http.ResponseWriter, r *http.Request) {
	if _, err := h.extractAdminClaims(r); err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	configs, err := h.repo.ListWhatsAppTemplateConfigs(r.Context())
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list templates", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"templates": configs})
}

// 8. PUT /v1/notification/admin/whatsapp-templates/{key}
func (h *AdminHandler) UpdateWhatsAppTemplate(w http.ResponseWriter, r *http.Request) {
	if _, err := h.extractAdminClaims(r); err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	var req struct {
		WhatsAppTemplateName string `json:"whatsapp_template_name"`
		IsApproved           bool   `json:"is_approved"`
		Language             string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}

	if req.Language == "" {
		req.Language = "en"
	}

	config := &WhatsAppTemplateConfig{
		TemplateKey:          key,
		WhatsAppTemplateName: req.WhatsAppTemplateName,
		IsApproved:           req.IsApproved,
		Language:             req.Language,
	}

	if err := h.repo.UpsertWhatsAppTemplateConfig(r.Context(), config); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update template", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

// 9. GET /v1/notification/admin/ops-alert-channels
func (h *AdminHandler) ListOpsAlertChannels(w http.ResponseWriter, r *http.Request) {
	if _, err := h.extractAdminClaims(r); err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	channels, err := h.repo.ListAllOpsAlertChannels(r.Context())
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list channels", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"channels": channels})
}

// 10. POST /v1/notification/admin/ops-alert-channels
func (h *AdminHandler) CreateOpsAlertChannel(w http.ResponseWriter, r *http.Request) {
	if _, err := h.extractAdminClaims(r); err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var ch OpsAlertChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil || ch.Target == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}

	if err := h.repo.CreateOpsAlertChannel(r.Context(), &ch); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to create channel", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ch)
}

// 11. PUT /v1/notification/admin/ops-alert-channels/{id}
func (h *AdminHandler) UpdateOpsAlertChannel(w http.ResponseWriter, r *http.Request) {
	if _, err := h.extractAdminClaims(r); err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var ch OpsAlertChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}
	ch.ID = id

	if err := h.repo.UpdateOpsAlertChannel(r.Context(), &ch); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update channel", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ch)
}
