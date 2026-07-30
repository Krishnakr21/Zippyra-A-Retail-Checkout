package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stdErrors "errors"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/crypto"
	sharedErrors "github.com/zippyra/backend/shared/errors"
)

type DownstreamClient interface {
	UpdateProductPrice(ctx *http.Request, barcode, storeID string, pricePaise int64) error
	AdjustInventory(ctx *http.Request, storeID, barcode, reason string, qty int) error
	ReceiveGRN(ctx *http.Request, storeID string, items []map[string]interface{}) error
}

type MockDownstreamClient struct {
	PriceUpdateCalls    int
	StockAdjustmentCalls int
	GRNCalls            int
}

func (m *MockDownstreamClient) UpdateProductPrice(r *http.Request, barcode, storeID string, pricePaise int64) error {
	m.PriceUpdateCalls++
	return nil
}

func (m *MockDownstreamClient) AdjustInventory(r *http.Request, storeID, barcode, reason string, qty int) error {
	m.StockAdjustmentCalls++
	return nil
}

func (m *MockDownstreamClient) ReceiveGRN(r *http.Request, storeID string, items []map[string]interface{}) error {
	m.GRNCalls++
	return nil
}

type WebhookHandler struct {
	repo       IntegrationRepository
	masterKey  string
	downstream DownstreamClient
}

func NewWebhookHandler(repo IntegrationRepository, masterKey string, downstream DownstreamClient) *WebhookHandler {
	if downstream == nil {
		downstream = &MockDownstreamClient{}
	}
	return &WebhookHandler{
		repo:       repo,
		masterKey:  masterKey,
		downstream: downstream,
	}
}

// 7. POST /v1/integration/connections/{id}/webhook
func (h *WebhookHandler) HandleInboundWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connID := vars["id"]

	conn, err := h.repo.GetConnectionByID(r.Context(), connID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Connection not found", nil)
		return
	}

	// 1. Read Raw Body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Failed to read body", nil)
		return
	}

	// 2. Signature Verification (HMAC-SHA256) BEFORE parsing JSON
	signatureHeader := r.Header.Get("X-Signature")
	if signatureHeader == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, "INVALID_SIGNATURE", "Missing X-Signature header", nil)
		return
	}

	secretBytes, err := crypto.Decrypt(conn.InboundWebhookSecretEncrypted, h.masterKey)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to decrypt webhook secret", nil)
		return
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write(bodyBytes)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signatureHeader), []byte(expectedSig)) {
		sharedErrors.WriteError(w, http.StatusBadRequest, "INVALID_SIGNATURE", "HMAC signature mismatch", nil)
		return
	}

	// 3. Compute Event ID (Idempotency Key)
	eventID := r.Header.Get("X-Event-Id")
	if eventID == "" {
		hSum := sha256.Sum256(bodyBytes)
		eventID = hex.EncodeToString(hSum[:])
	}

	var payload struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	// 4. Record Event (Idempotency DO NOTHING on conflict)
	event := &ERPWebhookEvent{
		ConnectionID:     connID,
		EventID:          eventID,
		EventType:        payload.EventType,
		RawPayload:       bodyBytes,
		ProcessingResult: ProcessingResultPending,
	}

	created, err := h.repo.CreateWebhookEvent(r.Context(), event)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to record webhook event", nil)
		return
	}

	// Gap #68: 0 rows affected -> already processed, return 200 immediately
	if !created {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "DUPLICATE_IGNORED",
			"event_id": eventID,
		})
		return
	}

	// 5. Downstream Routing by event_type
	var procErr error
	switch payload.EventType {
	case "PRICE_UPDATE":
		var pData struct {
			Barcode    string `json:"barcode"`
			StoreID    string `json:"store_id"`
			PricePaise int64  `json:"price_paise"`
		}
		_ = json.Unmarshal(payload.Payload, &pData)
		procErr = h.downstream.UpdateProductPrice(r, pData.Barcode, pData.StoreID, pData.PricePaise)

	case "STOCK_ADJUSTMENT":
		var sData struct {
			StoreID  string `json:"store_id"`
			Barcode  string `json:"barcode"`
			Reason   string `json:"reason"`
			Quantity int    `json:"quantity"`
		}
		_ = json.Unmarshal(payload.Payload, &sData)
		reason := sData.Reason
		if reason == "" {
			reason = "ADMIN_ERROR"
		}
		procErr = h.downstream.AdjustInventory(r, sData.StoreID, sData.Barcode, reason, sData.Quantity)

	case "GRN_RECEIVED":
		var gData struct {
			StoreID string                   `json:"store_id"`
			Items   []map[string]interface{} `json:"items"`
		}
		_ = json.Unmarshal(payload.Payload, &gData)
		procErr = h.downstream.ReceiveGRN(r, gData.StoreID, gData.Items)

	default:
		procErr = stdErrors.New("UNSUPPORTED_EVENT_TYPE")
	}

	// 6. Update Webhook Processing Result
	result := ProcessingResultApplied
	var failReason *string
	if procErr != nil {
		result = ProcessingResultFailed
		msg := procErr.Error()
		failReason = &msg
	}

	_ = h.repo.UpdateWebhookEventResult(r.Context(), event.ID, result, failReason)

	// Update Connection Timestamps & Status
	now := time.Now()
	var newStatus *ConnectionStatus
	if conn.Status == ConnectionStatusPendingSetup && result == ProcessingResultApplied {
		active := ConnectionStatusActive
		newStatus = &active
	}
	_ = h.repo.UpdateConnectionTimestamps(r.Context(), connID, &now, nil, nil, newStatus)

	// Return 200 to caller
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            string(result),
		"event_id":          eventID,
		"processing_result": result,
	})
}

func ComputeHMACSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
