package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zippyra/backend/shared/jwt"
)

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.1.0", -1},
		{"1.2.0", "1.2.0", 0},
		{"2.0.0", "1.9.9", 1},
		{"v1.0.0", "1.0.0", 0},
		{"V2.1.3", "2.1.2", 1},
		{"1.0.0-beta", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"0.9.0", "1.0.0", -1},
	}

	for _, tt := range tests {
		result := CompareSemver(tt.v1, tt.v2)
		assert.Equal(t, tt.expected, result, "CompareSemver(%s, %s)", tt.v1, tt.v2)
	}
}

func TestVersionCheck_HardUpdateRequired(t *testing.T) {
	repo := NewMemoryRepository()
	softMsg := "Please upgrade for new features"
	err := repo.UpsertAppVersion(nil, &AppVersion{
		Platform:            "ANDROID",
		MinSupportedVersion: "1.1.0",
		LatestVersion:       "1.5.0",
		HardUpdateBelow:     "1.2.0",
		SoftUpdateMessage:   &softMsg,
	})
	assert.NoError(t, err)

	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	// current_version (1.1.5) < hard_update_below (1.2.0) -> update_required = true
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/version-check?platform=ANDROID&current_version=1.1.5", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp VersionCheckResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.UpdateRequired)
	assert.Equal(t, "1.5.0", resp.LatestVersion)
	assert.Equal(t, "1.1.0", resp.MinSupportedVersion)
}

func TestVersionCheck_SoftUpdateMessage(t *testing.T) {
	repo := NewMemoryRepository()
	softMsg := "A new performance update is available!"
	err := repo.UpsertAppVersion(nil, &AppVersion{
		Platform:            "IOS",
		MinSupportedVersion: "1.0.0",
		LatestVersion:       "2.0.0",
		HardUpdateBelow:     "1.1.0",
		SoftUpdateMessage:   &softMsg,
	})
	assert.NoError(t, err)

	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	// min_supported (1.0.0) <= current (1.5.0) < latest (2.0.0) -> update_required = false, soft_update_message present
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/version-check?platform=IOS&current_version=1.5.0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp VersionCheckResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.UpdateRequired)
	assert.Equal(t, "2.0.0", resp.LatestVersion)
	assert.NotNil(t, resp.SoftUpdateMessage)
	assert.Equal(t, softMsg, *resp.SoftUpdateMessage)
}

func TestVersionCheck_AlreadyOnLatest(t *testing.T) {
	repo := NewMemoryRepository()
	softMsg := "New update available"
	err := repo.UpsertAppVersion(nil, &AppVersion{
		Platform:            "ANDROID",
		MinSupportedVersion: "1.0.0",
		LatestVersion:       "2.0.0",
		HardUpdateBelow:     "1.1.0",
		SoftUpdateMessage:   &softMsg,
	})
	assert.NoError(t, err)

	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	// current_version (2.0.0) >= latest (2.0.0) -> update_required = false, no soft message
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/version-check?platform=ANDROID&current_version=2.0.0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp VersionCheckResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.UpdateRequired)
	assert.Nil(t, resp.SoftUpdateMessage)
}

func TestVersionCheck_InvalidPlatform(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/version-check?platform=WINDOWS&current_version=1.0.0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	errObj := errResp["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_PLATFORM", errObj["code"])
}

func TestAdminUpdateAppVersion_RequiresStepUp(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	secret := "zippyra-dev-jwt-secret-key-32bytes"
	softMsg := "Critical security update"
	updateReq := UpdateAppVersionRequest{
		MinSupportedVersion: "1.2.0",
		LatestVersion:       "2.1.0",
		HardUpdateBelow:     "1.3.0",
		SoftUpdateMessage:   &softMsg,
	}
	bodyBytes, _ := json.Marshal(updateReq)

	// 1. Request without Authorization header -> 401 Unauthorized
	req1 := httptest.NewRequest(http.MethodPut, "/v1/auth/admin/app-versions/ANDROID", bytes.NewReader(bodyBytes))
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusUnauthorized, rr1.Code)

	// 2. Request with stale step-up token (> 10 min ago) -> 403 Forbidden
	staleClaims := &jwt.Claims{
		UserID:   "admin-123",
		UserType: "ADMIN",
		StepUpAt: time.Now().Add(-15 * time.Minute).Unix(),
	}
	staleToken, err := jwt.GenerateToken(staleClaims, secret, 1*time.Hour)
	assert.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodPut, "/v1/auth/admin/app-versions/ANDROID", bytes.NewReader(bodyBytes))
	req2.Header.Set("Authorization", "Bearer "+staleToken)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusForbidden, rr2.Code)

	var errResp2 map[string]interface{}
	_ = json.Unmarshal(rr2.Body.Bytes(), &errResp2)
	errObj2 := errResp2["error"].(map[string]interface{})
	assert.Equal(t, "STEP_UP_REQUIRED", errObj2["code"])

	// 3. Request with fresh step-up token (< 10 min ago) -> 200 OK
	freshClaims := &jwt.Claims{
		UserID:   "admin-123",
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}
	freshToken, err := jwt.GenerateToken(freshClaims, secret, 1*time.Hour)
	assert.NoError(t, err)

	req3 := httptest.NewRequest(http.MethodPut, "/v1/auth/admin/app-versions/ANDROID", bytes.NewReader(bodyBytes))
	req3.Header.Set("Authorization", "Bearer "+freshToken)
	rr3 := httptest.NewRecorder()
	router.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)

	var updated AppVersion
	err = json.Unmarshal(rr3.Body.Bytes(), &updated)
	assert.NoError(t, err)
	assert.Equal(t, "ANDROID", updated.Platform)
	assert.Equal(t, "1.2.0", updated.MinSupportedVersion)
	assert.Equal(t, "2.1.0", updated.LatestVersion)
	assert.Equal(t, "1.3.0", updated.HardUpdateBelow)

	// Verify persistence in repository
	av, err := repo.GetAppVersion(nil, "ANDROID")
	assert.NoError(t, err)
	assert.NotNil(t, av)
	assert.Equal(t, "2.1.0", av.LatestVersion)
}
