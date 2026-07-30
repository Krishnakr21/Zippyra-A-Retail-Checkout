package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

// PUT /v1/auth/admin/app-versions/{platform}
func (h *AuthHandler) HandleUpdateAppVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		errors.WriteError(w, http.StatusBadRequest, "INVALID_PLATFORM", "Missing platform parameter in URL path", nil)
		return
	}

	platformParam := strings.TrimSpace(parts[4])
	platform := strings.ToUpper(platformParam)
	if platform != "ANDROID" && platform != "IOS" {
		errors.WriteError(w, http.StatusBadRequest, "INVALID_PLATFORM", "Invalid platform. Allowed values: ANDROID, IOS", nil)
		return
	}

	var req UpdateAppVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid request body", nil)
		return
	}

	req.MinSupportedVersion = strings.TrimSpace(req.MinSupportedVersion)
	req.LatestVersion = strings.TrimSpace(req.LatestVersion)
	req.HardUpdateBelow = strings.TrimSpace(req.HardUpdateBelow)

	if req.MinSupportedVersion == "" || req.LatestVersion == "" || req.HardUpdateBelow == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "min_supported_version, latest_version, and hard_update_below are required", nil)
		return
	}

	av := AppVersion{
		Platform:            platform,
		MinSupportedVersion: req.MinSupportedVersion,
		LatestVersion:       req.LatestVersion,
		HardUpdateBelow:     req.HardUpdateBelow,
		SoftUpdateMessage:   req.SoftUpdateMessage,
	}

	ctx := r.Context()
	if err := h.repo.UpsertAppVersion(ctx, &av); err != nil {
		logger.Error("Failed to upsert app version for %s: %v", platform, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update app version", nil)
		return
	}

	// Invalidate / update Redis cache
	cacheKey := "app_version:" + platform
	if h.redisClient != nil && h.redisClient.Client != nil && h.redisClient.Ping(ctx).Err() == nil {
		if bytes, err := json.Marshal(&av); err == nil {
			_ = h.redisClient.Set(ctx, cacheKey, string(bytes), 5*time.Minute).Err()
		} else {
			_ = h.redisClient.Del(ctx, cacheKey).Err()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(av)
}
