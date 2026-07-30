package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type DeviceHandler struct {
	repo      Repository
	iot       IoTProvider
	redisClient *redis.Client
	auditPub  *audit.Publisher
	jwtSecret string
}

func NewDeviceHandler(repo Repository, iot IoTProvider, redisClient *redis.Client, auditPub *audit.Publisher) *DeviceHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &DeviceHandler{
		repo:        repo,
		iot:         iot,
		redisClient: redisClient,
		auditPub:    auditPub,
		jwtSecret:   secret,
	}
}

func (h *DeviceHandler) getClaims(r *http.Request) *jwt.Claims {
	if val := r.Context().Value("user_claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	if val := r.Context().Value("claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	return nil
}

func (h *DeviceHandler) HandleProvisionDevice(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required for device provisioning", nil)
		return
	}

	var req ProvisionDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoreID == "" || req.DeviceType == "" || req.Label == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, device_type, and label are required", nil)
		return
	}

	req.DeviceType = strings.ToUpper(strings.TrimSpace(req.DeviceType))

	if req.DeviceType == DeviceTypeGate {
		if req.GateID == nil || strings.TrimSpace(*req.GateID) == "" {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "gate_id is required for GATE device_type", nil)
			return
		}
		// Unique gate_id per store validation
		existingGateDevice, _ := h.repo.GetDeviceByGateID(r.Context(), req.StoreID, *req.GateID)
		if existingGateDevice != nil {
			errors.WriteError(w, http.StatusConflict, CodeGateIDAlreadyExists, "A device with this gate_id already exists for this store", nil)
			return
		}
	}

	deviceID := uuid.New().String()
	thingName := fmt.Sprintf("zippyra-thing-%s-%s", req.DeviceType, deviceID[:8])
	chainID := req.ChainID
	if chainID == "" {
		chainID = "chain-default-1"
	}

	// Step 1: Provision AWS IoT Thing + Certificate
	iotBundle, err := h.iot.CreateThingAndCert(r.Context(), thingName, req.DeviceType, req.StoreID, deviceID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, CodeProvisioningFailed, "Failed to provision AWS IoT resources: "+err.Error(), nil)
		return
	}

	device := &Device{
		ID:            deviceID,
		StoreID:       req.StoreID,
		ChainID:       chainID,
		DeviceType:    req.DeviceType,
		GateID:        req.GateID,
		Label:         req.Label,
		Status:        StatusProvisioning,
		IoTThingName:  thingName,
		CertARN:       iotBundle.CertARN,
		CertID:        iotBundle.CertID,
		CertExpiresAt: &iotBundle.ExpiresAt,
	}

	// Step 2: Issue 1-year DEVICE JWT
	deviceJWT, kid, err := IssueDeviceJWT(device, h.jwtSecret)
	if err != nil {
		// Rollback AWS IoT Thing on JWT creation error
		_ = h.iot.DecommissionThingAndCert(r.Context(), thingName, iotBundle.CertID)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to mint DEVICE JWT", nil)
		return
	}
	device.DeviceJWTKid = kid

	// Step 3: Insert DB record
	if err := h.repo.CreateDevice(r.Context(), device); err != nil {
		_ = h.iot.DecommissionThingAndCert(r.Context(), thingName, iotBundle.CertID)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to save device record", nil)
		return
	}

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.AdminID,
			ActionType:    "device.provisioned",
			TargetType:    "device",
			TargetID:      device.ID,
			SourceService: "device-mgmt-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"store_id": req.StoreID, "device_type": req.DeviceType, "label": req.Label},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProvisionDeviceResponse{
		DeviceID:      device.ID,
		DeviceJWT:     deviceJWT,
		CertPEM:       iotBundle.CertPEM,
		PrivateKeyPEM: iotBundle.PrivateKeyPEM,
		RootCAPEM:     iotBundle.RootCAPEM,
		MQTTEndpoint:  iotBundle.MQTTEndpoint,
	})
}

func (h *DeviceHandler) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	storeID := r.URL.Query().Get("store_id")
	statusFilter := r.URL.Query().Get("status")
	deviceTypeFilter := r.URL.Query().Get("device_type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	devices, total, err := h.repo.ListDevices(r.Context(), storeID, statusFilter, deviceTypeFilter, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list devices", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices":   devices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *DeviceHandler) HandleGetDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	device, err := h.repo.GetDeviceByID(r.Context(), id)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, CodeDeviceNotFound, "Device not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) HandleDecommissionDevice(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required to decommission device", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	device, err := h.repo.GetDeviceByID(r.Context(), id)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, CodeDeviceNotFound, "Device not found", nil)
		return
	}

	if device.Status == StatusDecommissioned {
		errors.WriteError(w, http.StatusBadRequest, CodeDeviceAlreadyDecommissioned, "Device is already decommissioned", nil)
		return
	}

	// Deactivate and delete AWS IoT resources
	_ = h.iot.DecommissionThingAndCert(r.Context(), device.IoTThingName, device.CertID)
	_ = h.repo.DecommissionDevice(r.Context(), id)

	// Set Redis revocation key for 1 year matching DEVICE JWT max validity
	if h.redisClient != nil {
		revocationKey := fmt.Sprintf("device_revoked:%s", id)
		_ = h.redisClient.Set(context.Background(), revocationKey, "true", 365*24*time.Hour).Err()
	}

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.AdminID,
			ActionType:    "device.decommissioned",
			TargetType:    "device",
			TargetID:      id,
			SourceService: "device-mgmt-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"store_id": device.StoreID, "iot_thing_name": device.IoTThingName},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "DECOMMISSIONED", "device_id": id})
}

func (h *DeviceHandler) HandleRotateCert(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required to rotate device certificate", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	device, err := h.repo.GetDeviceByID(r.Context(), id)
	if err != nil || device.Status == StatusDecommissioned {
		errors.WriteError(w, http.StatusNotFound, CodeDeviceNotFound, "Active device not found", nil)
		return
	}

	iotBundle, err := h.iot.RotateCert(r.Context(), device.IoTThingName, device.CertID, device.DeviceType, device.StoreID, device.ID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to rotate certificate", nil)
		return
	}

	deviceJWT, _, _ := IssueDeviceJWT(device, h.jwtSecret)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProvisionDeviceResponse{
		DeviceID:      device.ID,
		DeviceJWT:     deviceJWT,
		CertPEM:       iotBundle.CertPEM,
		PrivateKeyPEM: iotBundle.PrivateKeyPEM,
		RootCAPEM:     iotBundle.RootCAPEM,
		MQTTEndpoint:  iotBundle.MQTTEndpoint,
	})
}

func (h *DeviceHandler) HandleGetHeartbeats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}

	heartbeats, err := h.repo.GetHeartbeatsHypertable(r.Context(), id, from, to)
	if err != nil {
		heartbeats = []*DeviceHeartbeat{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"heartbeats": heartbeats})
}

func (h *DeviceHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	storeID := r.URL.Query().Get("store_id")
	resolvedStr := r.URL.Query().Get("resolved")

	var resolved *bool
	if resolvedStr != "" {
		val := resolvedStr == "true"
		resolved = &val
	}

	alerts, err := h.repo.ListAlerts(r.Context(), storeID, resolved)
	if err != nil {
		alerts = []*DeviceAlert{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"alerts": alerts})
}

func (h *DeviceHandler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.repo.ResolveAlertByID(r.Context(), id); err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Alert not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "RESOLVED", "alert_id": id})
}

var (
	inMemPairingMu    sync.Mutex
	inMemPairingStore = make(map[string]inMemPairingItem)
)

type inMemPairingItem struct {
	payload   []byte
	expiresAt time.Time
}

type GeneratePairingCodeResponse struct {
	PairingCode string `json:"pairing_code"`
}

type PairDeviceRequest struct {
	PairingCode string `json:"pairing_code"`
}

func (h *DeviceHandler) generateRandomCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func (h *DeviceHandler) HandleGeneratePairingCode(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required to generate pairing code", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	device, err := h.repo.GetDeviceByID(r.Context(), id)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, CodeDeviceNotFound, "Device not found", nil)
		return
	}

	// Generate 8-char code
	code := h.generateRandomCode(8)
	codeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))

	// Create/issue IoT credentials bundle for pairing
	iotBundle, err := h.iot.CreateThingAndCert(r.Context(), device.IoTThingName, device.DeviceType, device.StoreID, device.ID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, CodeProvisioningFailed, "Failed to prepare IoT credentials: "+err.Error(), nil)
		return
	}

	deviceJWT, _, err := IssueDeviceJWT(device, h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue DEVICE JWT", nil)
		return
	}

	respPayload := ProvisionDeviceResponse{
		DeviceID:      device.ID,
		DeviceJWT:     deviceJWT,
		CertPEM:       iotBundle.CertPEM,
		PrivateKeyPEM: iotBundle.PrivateKeyPEM,
		RootCAPEM:     iotBundle.RootCAPEM,
		MQTTEndpoint:  iotBundle.MQTTEndpoint,
	}

	payloadBytes, _ := json.Marshal(respPayload)
	redisKey := fmt.Sprintf("device_pairing:%s", codeHash)

	if h.redisClient != nil {
		h.redisClient.Set(r.Context(), redisKey, string(payloadBytes), 15*time.Minute)
	}

	inMemPairingMu.Lock()
	inMemPairingStore[codeHash] = inMemPairingItem{
		payload:   payloadBytes,
		expiresAt: time.Now().Add(15 * time.Minute),
	}
	inMemPairingMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GeneratePairingCodeResponse{
		PairingCode: code,
	})
}

func (h *DeviceHandler) HandlePairDevice(w http.ResponseWriter, r *http.Request) {
	var req PairDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.PairingCode) == "" {
		errors.WriteError(w, http.StatusBadRequest, CodePairingCodeInvalid, "pairing_code is required", nil)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.PairingCode))
	codeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))

	var payloadBytes []byte

	redisKey := fmt.Sprintf("device_pairing:%s", codeHash)
	if h.redisClient != nil {
		val, err := h.redisClient.Get(r.Context(), redisKey).Result()
		if err == nil && val != "" {
			payloadBytes = []byte(val)
			h.redisClient.Del(r.Context(), redisKey) // Single-use consumption
		}
	}

	if len(payloadBytes) == 0 {
		inMemPairingMu.Lock()
		item, ok := inMemPairingStore[codeHash]
		if ok {
			delete(inMemPairingStore, codeHash) // Single-use consumption
			if time.Now().Before(item.expiresAt) {
				payloadBytes = item.payload
			}
		}
		inMemPairingMu.Unlock()
	}

	if len(payloadBytes) == 0 {
		errors.WriteError(w, http.StatusBadRequest, CodePairingCodeInvalid, "Invalid or expired pairing code", nil)
		return
	}

	var resp ProvisionDeviceResponse
	if err := json.Unmarshal(payloadBytes, &resp); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to parse paired device credentials", nil)
		return
	}

	// Update device status to ACTIVE
	_ = h.repo.UpdateDeviceStatus(r.Context(), resp.DeviceID, StatusActive)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payloadBytes)
}

