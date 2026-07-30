package main

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type StoreHandler struct {
	repo        Repository
	capacityMgr CapacityManager
	sessionMgr  *SessionManager
	jwtSecret   string
}

func NewStoreHandler(repo Repository, capacityMgr CapacityManager, sessionMgr *SessionManager, jwtSecret string) *StoreHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &StoreHandler{
		repo:        repo,
		capacityMgr: capacityMgr,
		sessionMgr:  sessionMgr,
		jwtSecret:   jwtSecret,
	}
}

// GET /v1/store/nearby?lat={f}&lng={f}&radius_km={f}
func (h *StoreHandler) HandleGetNearbyStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	query := r.URL.Query()
	latStr := query.Get("lat")
	lngStr := query.Get("lng")
	radiusStr := query.Get("radius_km")

	if latStr == "" || lngStr == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "lat and lng parameters are required", nil)
		return
	}

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil || lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid lat or lng coordinates", nil)
		return
	}

	radiusKM := 10.0
	if radiusStr != "" {
		if rVal, err := strconv.ParseFloat(radiusStr, 64); err == nil && rVal > 0 {
			radiusKM = rVal
		}
	}
	if radiusKM > 25.0 {
		radiusKM = 25.0
	}

	stores, err := h.repo.GetNearbyStores(r.Context(), lat, lng, radiusKM)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query nearby stores", nil)
		return
	}

	now := time.Now()
	var resp []*NearbyStoreResponse
	for _, store := range stores {
		distKM := math.Round((HaversineDistance(lat, lng, store.Lat, store.Lng)/1000.0)*100) / 100
		liveCap, _ := h.capacityMgr.GetLiveCapacity(r.Context(), store.ID)
		capPct := 0
		if store.CapacityMax > 0 {
			capPct = int(math.Round((float64(liveCap) / float64(store.CapacityMax)) * 100))
			if capPct > 100 {
				capPct = 100
			}
		}

		isOpen := IsStoreOpenNow(store, now)

		resp = append(resp, &NearbyStoreResponse{
			ID:          store.ID,
			Name:        store.Name,
			Address:     store.Address,
			DistanceKM:  distKM,
			IsOpen:      isOpen,
			CapacityPct: capPct,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /v1/store/{store_id}
func (h *StoreHandler) HandleGetStoreDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[2] == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing store_id in path", nil)
		return
	}
	storeID := parts[2]

	store, err := h.repo.GetStoreByID(r.Context(), storeID)
	if err != nil || store == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeStoreNotFound, "Store not found", nil)
		return
	}

	liveCap, _ := h.capacityMgr.GetLiveCapacity(r.Context(), store.ID)
	capPct := 0
	if store.CapacityMax > 0 {
		capPct = int(math.Round((float64(liveCap) / float64(store.CapacityMax)) * 100))
		if capPct > 100 {
			capPct = 100
		}
	}

	isOpen := IsStoreOpenNow(store, time.Now())

	resp := &StoreDetailResponse{
		ID:                   store.ID,
		Name:                 store.Name,
		Address:              store.Address,
		City:                 store.City,
		State:                store.State,
		Pincode:              store.Pincode,
		Lat:                  store.Lat,
		Lng:                  store.Lng,
		OpeningTime:          store.OpeningTime,
		ClosingTime:          store.ClosingTime,
		Timezone:             store.Timezone,
		IsOpen:               isOpen,
		CapacityPct:          capPct,
		RFIDEnabled:          store.RFIDEnabled,
		GeofenceRadiusMeters: store.GeofenceRadiusMeters,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /v1/store/bind (CUSTOMER JWT)
func (h *StoreHandler) HandleBindStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID := h.getUserIDFromContext(r)
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	var req StoreBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.QRToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "qr_token is required", nil)
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	resp, err := h.sessionMgr.BindStore(r.Context(), userID, clientIP, &req)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			status := http.StatusBadRequest
			switch apiErr.Code {
			case errors.CodeRateLimitExceeded:
				status = http.StatusTooManyRequests
			case errors.CodeStoreClosed, errors.CodeStoreAtCapacity:
				status = http.StatusConflict
			case errors.CodeStoreGeofenceMismatch:
				status = http.StatusForbidden
			}
			errors.WriteError(w, status, apiErr.Code, apiErr.Message, apiErr.Details)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /v1/store/unbind (CUSTOMER JWT or Session JWT)
func (h *StoreHandler) HandleUnbindStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID := h.getUserIDFromContext(r)
	var req StoreUnbindRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if userID == "" && req.SessionID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication or session_id required", nil)
		return
	}

	if err := h.sessionMgr.UnbindStore(r.Context(), userID, req.SessionID, "customer_exit"); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to unbind session", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UNBOUND"})
}

// GET /v1/store/session (CUSTOMER JWT)
func (h *StoreHandler) HandleGetActiveSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID := h.getUserIDFromContext(r)
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	resp, err := h.sessionMgr.GetActiveSession(r.Context(), userID)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			errors.WriteError(w, http.StatusNotFound, apiErr.Code, apiErr.Message, nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *StoreHandler) getUserIDFromContext(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		// Try Session token parse
		sessClaims, errSess := jwt.ParseAndVerifySessionToken(tokenStr, h.jwtSecret)
		if errSess == nil && sessClaims != nil {
			return sessClaims.UserID
		}
		return ""
	}
	return claims.UserID
}

// GET /v1/store/home-banners (Public)
func (h *StoreHandler) HandleGetHomeBanners(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	banners := []*HomeBanner{
		{
			ID:       "banner-1",
			Title:    "Welcome to Zippyra Scan & Go",
			ImageURL: "https://images.unsplash.com/photo-1604719312566-8912e9227c6a?w=800&auto=format&fit=crop",
			DeepLink: "/store/scan",
		},
		{
			ID:       "banner-2",
			Title:    "Earn Double Loyalty Points This Weekend",
			ImageURL: "https://images.unsplash.com/photo-1556742049-0a670f4a4591?w=800&auto=format&fit=crop",
			DeepLink: "/loyalty",
		},
		{
			ID:       "banner-3",
			Title:    "Express Self Checkout at Nearby Stores",
			ImageURL: "https://images.unsplash.com/photo-1578916171728-46686eac8d58?w=800&auto=format&fit=crop",
			DeepLink: "/store/list",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HomeBannersResponse{Banners: banners})
}
