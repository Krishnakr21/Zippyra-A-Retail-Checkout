package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

// GET /v1/auth/version-check?platform={ANDROID|IOS}&current_version={semver}
func (h *AuthHandler) HandleVersionCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	platformParam := strings.TrimSpace(r.URL.Query().Get("platform"))
	platform := strings.ToUpper(platformParam)
	if platform != "ANDROID" && platform != "IOS" {
		errors.WriteError(w, http.StatusBadRequest, "INVALID_PLATFORM", "Invalid or missing platform. Allowed values: ANDROID, IOS", nil)
		return
	}

	currentVersion := strings.TrimSpace(r.URL.Query().Get("current_version"))
	if currentVersion == "" {
		currentVersion = "0.0.0"
	}

	ctx := r.Context()
	cacheKey := "app_version:" + platform

	var av *AppVersion

	// 1. Try Redis cache
	if h.redisClient != nil && h.redisClient.Client != nil && h.redisClient.Ping(ctx).Err() == nil {
		if val, err := h.redisClient.Get(ctx, cacheKey).Result(); err == nil && val != "" {
			var cached AppVersion
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				av = &cached
			}
		}
	}

	// 2. Fetch from DB on cache miss
	if av == nil {
		var err error
		av, err = h.repo.GetAppVersion(ctx, platform)
		if err != nil {
			logger.Error("Failed to fetch app version for %s: %v", platform, err)
		}

		if av == nil {
			// Default fallback if table record doesn't exist yet
			av = &AppVersion{
				Platform:            platform,
				MinSupportedVersion: "1.0.0",
				LatestVersion:       "1.0.0",
				HardUpdateBelow:     "1.0.0",
			}
		} else {
			// Populate Redis cache (5 min TTL)
			if h.redisClient != nil && h.redisClient.Client != nil && h.redisClient.Ping(ctx).Err() == nil {
				if bytes, err := json.Marshal(av); err == nil {
					_ = h.redisClient.Set(ctx, cacheKey, string(bytes), 5*time.Minute).Err()
				}
			}
		}
	}

	// 3. Version comparison logic
	// update_required = true if current_version < hard_update_below
	updateRequired := CompareSemver(currentVersion, av.HardUpdateBelow) < 0

	var softMsg *string
	if !updateRequired {
		// If min_supported_version <= current_version < latest_version: include soft_update_message
		if CompareSemver(av.MinSupportedVersion, currentVersion) <= 0 && CompareSemver(currentVersion, av.LatestVersion) < 0 {
			if av.SoftUpdateMessage != nil && *av.SoftUpdateMessage != "" {
				softMsg = av.SoftUpdateMessage
			}
		}
	}

	resp := VersionCheckResponse{
		UpdateRequired:      updateRequired,
		LatestVersion:       av.LatestVersion,
		MinSupportedVersion: av.MinSupportedVersion,
		SoftUpdateMessage:   softMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// CompareSemver compares two semver strings (e.g. "1.2.0", "v1.2.0", "1.0.0-beta").
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func CompareSemver(v1, v2 string) int {
	clean1, meta1 := parseSemver(v1)
	clean2, meta2 := parseSemver(v2)

	for i := 0; i < 3; i++ {
		if clean1[i] < clean2[i] {
			return -1
		}
		if clean1[i] > clean2[i] {
			return 1
		}
	}

	// Main versions are equal; compare prerelease metadata
	if meta1 == "" && meta2 != "" {
		return 1 // release version is greater than prerelease
	}
	if meta1 != "" && meta2 == "" {
		return -1 // prerelease is smaller than release
	}
	if meta1 < meta2 {
		return -1
	}
	if meta1 > meta2 {
		return 1
	}
	return 0
}

func parseSemver(v string) ([3]int, string) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")

	var meta string
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		meta = v[idx:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		num, err := strconv.Atoi(parts[i])
		if err == nil {
			result[i] = num
		}
	}
	return result, meta
}
