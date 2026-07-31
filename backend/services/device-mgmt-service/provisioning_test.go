package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func TestProvisionDevice_Success_ReturnsCredentialsAndJWT(t *testing.T) {
	repo := NewMemoryRepository()
	iot := NewMockIoTProvider()
	handler := NewDeviceHandler(repo, iot, nil, nil)

	adminClaims := &jwt.Claims{
		AdminID:  "admin-super-1",
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	gateID := "GATE_01"
	body, _ := json.Marshal(ProvisionDeviceRequest{
		StoreID:    "store-100",
		DeviceType: DeviceTypeGate,
		GateID:     &gateID,
		Label:      "Main Store Gate",
	})

	req := httptest.NewRequest("POST", "/v1/device-mgmt/devices", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", adminClaims))
	w := httptest.NewRecorder()

	handler.HandleProvisionDevice(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProvisionDeviceResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.DeviceID == "" || resp.DeviceJWT == "" || resp.PrivateKeyPEM == "" {
		t.Fatalf("expected valid device_id, device_jwt, and private_key_pem in one-time bundle")
	}

	device, err := repo.GetDeviceByID(context.Background(), resp.DeviceID)
	if err != nil || device.Status != StatusProvisioning {
		t.Fatalf("expected device in PROVISIONING status, got %v, err: %v", device, err)
	}
}

func TestProvisionDevice_SimulatedAWSTimeoutFailure_RollbackNoOrphanDevice(t *testing.T) {
	repo := NewMemoryRepository()
	iot := NewMockIoTProvider()
	handler := NewDeviceHandler(repo, iot, nil, nil)

	adminClaims := &jwt.Claims{
		AdminID:  "admin-super-1",
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	body, _ := json.Marshal(ProvisionDeviceRequest{
		StoreID:    "store-100",
		DeviceType: "FAIL_TRIGGER",
		Label:      "fail-trigger",
	})

	req := httptest.NewRequest("POST", "/v1/device-mgmt/devices", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", adminClaims))
	w := httptest.NewRecorder()

	handler.HandleProvisionDevice(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on AWS failure, got %d", w.Code)
	}

	devices, total, _ := repo.ListDevices(context.Background(), "store-100", "", "", 1, 10)
	if total != 0 || len(devices) != 0 {
		t.Fatalf("expected 0 orphaned device records after AWS failure, got %d", total)
	}
}
