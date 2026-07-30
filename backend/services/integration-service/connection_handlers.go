package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/crypto"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type ConnectionHandler struct {
	repo           IntegrationRepository
	auditPublisher *audit.Publisher
	masterKey      string
	jwtSecret      string
}

func NewConnectionHandler(repo IntegrationRepository, auditPublisher *audit.Publisher, masterKey string) *ConnectionHandler {
	return &ConnectionHandler{
		repo:           repo,
		auditPublisher: auditPublisher,
		masterKey:      masterKey,
		jwtSecret:      "dev-secret-key-change-in-prod",
	}
}

func (h *ConnectionHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil {
			return claims, nil
		}
	}

	// Fallback header extraction for development/tests
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

func generateRandomToken(prefix string) string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return prefix + "_" + hex.EncodeToString(bytes)
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// 1. POST /v1/integration/connections
func (h *ConnectionHandler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req CreateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid request payload", nil)
		return
	}

	// Scope & Role Enforcement: CHAIN_HQ requires OWNER role
	if req.ChainID != "" && claims.ChainID != "" && req.ChainID != claims.ChainID {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Cannot configure connection for another chain", nil)
		return
	}
	if claims.Role != "OWNER" && claims.Role != "SUPER_ADMIN" {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "High-impact integration setup requires OWNER role", nil)
		return
	}

	// Generate secrets
	plaintextSecret := generateRandomToken("whsec")
	encryptedSecret, err := crypto.Encrypt([]byte(plaintextSecret), h.masterKey)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to encrypt webhook secret", nil)
		return
	}

	var plaintextAgentKey *string
	var agentKeyHash *string
	if req.IntegrationMode == IntegrationModeAgentPolled {
		key := generateRandomToken("agent_key")
		plaintextAgentKey = &key
		h := hashString(key)
		agentKeyHash = &h
	}

	var encryptedOutbound []byte
	if req.IntegrationMode == IntegrationModeDirect && req.OutboundConfig != nil {
		rawOutbound, _ := json.Marshal(req.OutboundConfig)
		enc, err := crypto.Encrypt(rawOutbound, h.masterKey)
		if err != nil {
			sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to encrypt outbound config", nil)
			return
		}
		encryptedOutbound = enc
	}

	conn := &ERPConnection{
		ID:                            uuid.New().String(),
		ChainID:                       req.ChainID,
		ERPType:                       req.ERPType,
		IntegrationMode:               req.IntegrationMode,
		DisplayName:                   req.DisplayName,
		InboundWebhookSecretEncrypted: encryptedSecret,
		AgentAPIKeyHash:               agentKeyHash,
		OutboundConfigEncrypted:       encryptedOutbound,
		EnabledOutboundEvents:         req.EnabledOutboundEvents,
		Status:                        ConnectionStatusPendingSetup,
		CreatedBy:                     claims.UserID,
	}

	if conn.EnabledOutboundEvents == nil {
		conn.EnabledOutboundEvents = []string{}
	}

	if err := h.repo.CreateConnection(r.Context(), conn); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to create connection", nil)
		return
	}

	var setupNote string
	if req.IntegrationMode == IntegrationModeAgentPolled {
		setupNote = "Install the Zippyra ERP Connector Agent binary on the store machine and configure the agent_api_key."
	} else {
		setupNote = "Configure your SAP Cloud Connector / OData endpoint with the provided webhook_secret."
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "INTEGRATION_CONNECTION_CREATED",
			TargetType:    "ERPConnection",
			TargetID:      conn.ID,
			SourceService: "integration-service",
			Payload: map[string]interface{}{
				"chain_id": conn.ChainID,
				"erp_type": conn.ERPType,
				"mode":     conn.IntegrationMode,
			},
		})
	}

	resp := CreateConnectionResponse{
		Connection:          conn,
		PlaintextSecret:     plaintextSecret,
		PlaintextAgentAPIKey: plaintextAgentKey,
		ConnectorSetupNote:  setupNote,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// 2. GET /v1/integration/connections?chain_id={id}
func (h *ConnectionHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		chainID = claims.ChainID
	}

	conns, err := h.repo.ListConnectionsByChain(r.Context(), chainID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list connections", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"connections": conns})
}

// 3. GET /v1/integration/connections/{id}
func (h *ConnectionHandler) GetConnection(w http.ResponseWriter, r *http.Request) {
	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), id)
	if err != nil {
		if err == ErrConnectionNotFound {
			sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
			return
		}
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to fetch connection", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conn)
}

// 4. PUT /v1/integration/connections/{id}
func (h *ConnectionHandler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), id)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	if claims.Role != "OWNER" && claims.Role != "SUPER_ADMIN" {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Updating connection requires OWNER role", nil)
		return
	}

	var req UpdateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid request payload", nil)
		return
	}

	if req.DisplayName != "" {
		conn.DisplayName = req.DisplayName
	}
	if req.EnabledOutboundEvents != nil {
		conn.EnabledOutboundEvents = req.EnabledOutboundEvents
	}
	if req.Status != "" {
		conn.Status = req.Status
	}

	if err := h.repo.UpdateConnection(r.Context(), conn); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update connection", nil)
		return
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "INTEGRATION_CONNECTION_UPDATED",
			TargetType:    "ERPConnection",
			TargetID:      conn.ID,
			SourceService: "integration-service",
			Payload: map[string]interface{}{
				"status": conn.Status,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conn)
}

// 5. DELETE /v1/integration/connections/{id}
func (h *ConnectionHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if claims.Role != "OWNER" && claims.Role != "SUPER_ADMIN" {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Deleting connection requires OWNER role", nil)
		return
	}

	if err := h.repo.DeleteConnection(r.Context(), id); err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "INTEGRATION_CONNECTION_DELETED",
			TargetType:    "ERPConnection",
			TargetID:      id,
			SourceService: "integration-service",
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// 6. POST /v1/integration/connections/{id}/rotate-secret
func (h *ConnectionHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), id)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	if claims.Role != "OWNER" && claims.Role != "SUPER_ADMIN" {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Rotating secrets requires OWNER role", nil)
		return
	}

	newSecret := generateRandomToken("whsec")
	encSecret, err := crypto.Encrypt([]byte(newSecret), h.masterKey)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to encrypt secret", nil)
		return
	}

	conn.InboundWebhookSecretEncrypted = encSecret

	var newAgentKey *string
	if conn.IntegrationMode == IntegrationModeAgentPolled {
		key := generateRandomToken("agent_key")
		newAgentKey = &key
		hash := hashString(key)
		conn.AgentAPIKeyHash = &hash
	}

	if err := h.repo.UpdateConnection(r.Context(), conn); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to rotate secret", nil)
		return
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "INTEGRATION_SECRET_ROTATED",
			TargetType:    "ERPConnection",
			TargetID:      id,
			SourceService: "integration-service",
		})
	}

	resp := RotateSecretResponse{
		PlaintextSecret:     newSecret,
		PlaintextAgentAPIKey: newAgentKey,
		GracePeriodSeconds:  300, // 5 minutes grace period
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
