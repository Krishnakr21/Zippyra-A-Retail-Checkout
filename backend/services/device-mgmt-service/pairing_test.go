package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zippyra/backend/shared/jwt"
)

func setupPairingTestHandler() (*DeviceHandler, *MemoryRepository) {
	repo := NewMemoryRepository()
	iot := NewMockIoTProvider()
	handler := NewDeviceHandler(repo, iot, nil, nil)
	return handler, repo
}

func generateTestToken(userType, adminID string) string {
	secret := "zippyra-dev-jwt-secret-key-32bytes"
	claims := jwt.Claims{
		AdminID:  adminID,
		UserType: userType,
	}
	token, _ := jwt.GenerateToken(&claims, secret, 1*time.Hour)
	return token
}

func TestGeneratePairingCode_AdminOnly(t *testing.T) {
	handler, repo := setupPairingTestHandler()
	device := &Device{
		ID:         "dev-pairing-1",
		StoreID:    "store-1",
		DeviceType: DeviceTypeKiosk,
		Label:      "Front Desk Kiosk",
		Status:     StatusProvisioning,
	}
	_ = repo.CreateDevice(context.Background(), device)

	router := SetupRoutes(handler)

	// Staff/Cashier JWT attempt -> 403 Forbidden
	staffToken := generateTestToken("STAFF", "staff-123")
	req, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/dev-pairing-1/generate-pairing-code", nil)
	req.Header.Set("Authorization", "Bearer "+staffToken)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Admin JWT attempt -> 200 OK with 8-char pairing_code
	adminToken := generateTestToken("ADMIN", "admin-999")
	reqAdmin, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/dev-pairing-1/generate-pairing-code", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	rrAdmin := httptest.NewRecorder()
	router.ServeHTTP(rrAdmin, reqAdmin)

	assert.Equal(t, http.StatusOK, rrAdmin.Code)

	var resp GeneratePairingCodeResponse
	err := json.Unmarshal(rrAdmin.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.PairingCode, 8)
}

func TestPairDevice_Success_And_SingleUse(t *testing.T) {
	handler, repo := setupPairingTestHandler()
	device := &Device{
		ID:           "dev-pairing-2",
		StoreID:      "store-2",
		DeviceType:   DeviceTypeKiosk,
		Label:        "Self Service Kiosk",
		IoTThingName: "thing-kiosk-2",
		Status:       StatusProvisioning,
	}
	_ = repo.CreateDevice(context.Background(), device)

	router := SetupRoutes(handler)

	// 1. Admin generates code
	adminToken := generateTestToken("ADMIN", "admin-999")
	reqGen, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/dev-pairing-2/generate-pairing-code", nil)
	reqGen.Header.Set("Authorization", "Bearer "+adminToken)
	rrGen := httptest.NewRecorder()
	router.ServeHTTP(rrGen, reqGen)
	assert.Equal(t, http.StatusOK, rrGen.Code)

	var genResp GeneratePairingCodeResponse
	_ = json.Unmarshal(rrGen.Body.Bytes(), &genResp)
	code := genResp.PairingCode

	// 2. Unauthenticated staff device redeems pairing code
	pairBody, _ := json.Marshal(map[string]string{"pairing_code": code})
	reqPair, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/pair", bytes.NewBuffer(pairBody))
	rrPair := httptest.NewRecorder()
	router.ServeHTTP(rrPair, reqPair)

	assert.Equal(t, http.StatusOK, rrPair.Code)

	var creds ProvisionDeviceResponse
	err := json.Unmarshal(rrPair.Body.Bytes(), &creds)
	assert.NoError(t, err)
	assert.Equal(t, "dev-pairing-2", creds.DeviceID)
	assert.NotEmpty(t, creds.CertPEM)
	assert.NotEmpty(t, creds.PrivateKeyPEM)
	assert.NotEmpty(t, creds.DeviceJWT)

	// Verify device status updated to ACTIVE in DB
	updatedDevice, _ := repo.GetDeviceByID(context.Background(), "dev-pairing-2")
	assert.Equal(t, StatusActive, updatedDevice.Status)

	// 3. Second redemption attempt with same code fails (single-use)
	reqPair2, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/pair", bytes.NewBuffer(pairBody))
	rrPair2 := httptest.NewRecorder()
	router.ServeHTTP(rrPair2, reqPair2)

	assert.Equal(t, http.StatusBadRequest, rrPair2.Code)
	assert.Contains(t, rrPair2.Body.String(), CodePairingCodeInvalid)
}

func TestPairDevice_InvalidCode(t *testing.T) {
	handler, _ := setupPairingTestHandler()
	router := SetupRoutes(handler)

	pairBody, _ := json.Marshal(map[string]string{"pairing_code": "INVALID8"})
	reqPair, _ := http.NewRequest("POST", "/v1/device-mgmt/devices/pair", bytes.NewBuffer(pairBody))
	rrPair := httptest.NewRecorder()
	router.ServeHTTP(rrPair, reqPair)

	assert.Equal(t, http.StatusBadRequest, rrPair.Code)
	assert.Contains(t, rrPair.Body.String(), CodePairingCodeInvalid)
}
