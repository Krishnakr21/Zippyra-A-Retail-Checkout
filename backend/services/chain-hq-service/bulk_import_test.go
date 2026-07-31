package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/jwt"
)

func TestBulkImport_InitiateAndPollStatus_Success(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewBulkImportHandler(repo, nil)

	claims := &jwt.Claims{
		UserID:   "user-owner-1",
		ChainID:  "chain-100",
		Role:     RoleOwner,
		UserType: "CHAIN_HQ",
	}

	body, _ := json.Marshal(BulkImportRequest{
		Target:   "specific_stores",
		StoreIDs: []string{"store-001", "store-002"},
	})

	req := httptest.NewRequest("POST", "/v1/chain-hq/catalog/bulk-import", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
	w := httptest.NewRecorder()

	handler.HandleBulkImport(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202 Accepted, got %d", w.Code)
	}

	var initResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &initResp)
	jobID := initResp["job_id"].(string)

	// Poll status
	reqStatus := httptest.NewRequest("GET", "/v1/chain-hq/catalog/bulk-import/"+jobID+"/status", nil)
	reqStatus = reqStatus.WithContext(context.WithValue(reqStatus.Context(), "user_claims", claims))
	reqStatus = mux.SetURLVars(reqStatus, map[string]string{"id": jobID})
	wStatus := httptest.NewRecorder()

	handler.HandleGetBulkImportStatus(wStatus, reqStatus)

	if wStatus.Code != http.StatusOK {
		t.Fatalf("expected status 200 on status poll, got %d", wStatus.Code)
	}

	var statusResp map[string]interface{}
	_ = json.Unmarshal(wStatus.Body.Bytes(), &statusResp)
	if statusResp["summary"] == nil {
		t.Fatalf("expected summary in status response")
	}
}
