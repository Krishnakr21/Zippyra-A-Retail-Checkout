package main

import (
	"time"
)

const (
	ChainStatusActive    = "ACTIVE"
	ChainStatusSuspended = "SUSPENDED"

	CodeChainNotFound  = "CHAIN_NOT_FOUND"
	CodeChainSuspended = "CHAIN_SUSPENDED"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Chain struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	LegalEntityName    string    `json:"legal_entity_name,omitempty"`
	DefaultGstInPrefix string    `json:"default_gstin_prefix,omitempty"`
	Status             string    `json:"status"`
	StoreCount         int       `json:"store_count,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateChainRequest struct {
	Name               string `json:"name"`
	LegalEntityName    string `json:"legal_entity_name,omitempty"`
	DefaultGstInPrefix string `json:"default_gstin_prefix,omitempty"`
}

type UpdateChainRequest struct {
	Name               *string `json:"name,omitempty"`
	LegalEntityName    *string `json:"legal_entity_name,omitempty"`
	DefaultGstInPrefix *string `json:"default_gstin_prefix,omitempty"`
	Status             *string `json:"status,omitempty"`
}

type UpdateChainStatusRequest struct {
	Status string `json:"status"`
}

type UpdatePaymentSetupRequest struct {
	RazorpayAccountID string `json:"razorpay_account_id"`
	RazorpayKYCStatus string `json:"razorpay_kyc_status"`
	PaymentSetupNote  string `json:"payment_setup_note"`
}

type HomeBanner struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ImageURL    string `json:"image_url"`
	DeepLink    string `json:"deep_link"`
	ActiveFrom  string `json:"active_from,omitempty"`
	ActiveUntil string `json:"active_until,omitempty"`
}

type HomeBannersResponse struct {
	Banners []*HomeBanner `json:"banners"`
}

type Store struct {
	ID                   string     `json:"id"`
	ChainID              string     `json:"chain_id"`
	Name                 string     `json:"name"`
	Address              string     `json:"address"`
	City                 string     `json:"city"`
	State                string     `json:"state"`
	Pincode              string     `json:"pincode"`
	GSTIN                string     `json:"gstin,omitempty"`
	Lat                  float64    `json:"lat"`
	Lng                  float64    `json:"lng"`
	GeofencePolygon      []Point    `json:"geofence_polygon,omitempty"`
	GeofenceRadiusMeters int        `json:"geofence_radius_meters"`
	CapacityMax          int        `json:"capacity_max"`
	OpeningTime          string     `json:"opening_time"` // e.g. "08:00:00"
	ClosingTime          string     `json:"closing_time"` // e.g. "22:00:00"
	Timezone             string     `json:"timezone"`
	Status               string     `json:"status"` // ACTIVE | INACTIVE | UNDER_MAINTENANCE
	RFIDEnabled          bool       `json:"rfid_enabled"`
	CatalogVersion       int64      `json:"catalog_version"`
	RazorpayAccountID    string     `json:"razorpay_account_id,omitempty"`
	RazorpayKYCStatus    string     `json:"razorpay_kyc_status,omitempty"`
	PaymentSetupNote     string     `json:"payment_setup_note,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type StoreQRToken struct {
	ID        string    `json:"id"`
	StoreID   string    `json:"store_id"`
	GateID    string    `json:"gate_id"`
	Token     string    `json:"token"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreSession struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	StoreID              string     `json:"store_id"`
	DeviceID             string     `json:"device_id,omitempty"`
	BoundAt              time.Time  `json:"bound_at"`
	UnboundAt            *time.Time `json:"unbound_at,omitempty"`
	CatalogVersionAtBind int64      `json:"catalog_version_at_bind"`
}

type NearbyStoreResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	DistanceKM  float64 `json:"distance_km"`
	IsOpen      bool    `json:"is_open"`
	CapacityPct int     `json:"capacity_pct"`
}

type StoreDetailResponse struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Address              string  `json:"address"`
	City                 string  `json:"city"`
	State                string  `json:"state"`
	Pincode              string  `json:"pincode"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	OpeningTime          string  `json:"opening_time"`
	ClosingTime          string  `json:"closing_time"`
	Timezone             string  `json:"timezone"`
	IsOpen               bool    `json:"is_open"`
	CapacityPct          int     `json:"capacity_pct"`
	RFIDEnabled          bool    `json:"rfid_enabled"`
	GeofenceRadiusMeters int     `json:"geofence_radius_meters"`
}

type StoreBindRequest struct {
	QRToken  string  `json:"qr_token"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	DeviceID string  `json:"device_id"`
}

type StoreBindResponse struct {
	StoreID          string `json:"store_id"`
	StoreName        string `json:"store_name"`
	SessionToken     string `json:"session_token"`
	SessionExpiresAt string `json:"session_expires_at"`
	CatalogVersion   int64  `json:"catalog_version"`
	RFIDEnabled      bool   `json:"rfid_enabled"`
}

type StoreUnbindRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

type StoreSessionResponse struct {
	StoreID          string `json:"store_id"`
	StoreName        string `json:"store_name"`
	SessionToken     string `json:"session_token"`
	SessionExpiresAt string `json:"session_expires_at"`
	CatalogVersion   int64  `json:"catalog_version"`
}

type AdminStoreCreateRequest struct {
	ChainID              string  `json:"chain_id"`
	Name                 string  `json:"name"`
	Address              string  `json:"address"`
	City                 string  `json:"city"`
	State                string  `json:"state"`
	Pincode              string  `json:"pincode"`
	GSTIN                string  `json:"gstin,omitempty"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	GeofenceRadiusMeters int     `json:"geofence_radius_meters"`
	CapacityMax          int     `json:"capacity_max"`
	OpeningTime          string  `json:"opening_time"`
	ClosingTime          string  `json:"closing_time"`
	Timezone             string  `json:"timezone"`
	RFIDEnabled          bool    `json:"rfid_enabled"`
}

type AdminGeofenceUpdateRequest struct {
	Polygon      []Point `json:"polygon,omitempty"`
	RadiusMeters int     `json:"radius_meters"`
}

type AdminHoursUpdateRequest struct {
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
	Timezone    string `json:"timezone"`
}

type AdminCapacityUpdateRequest struct {
	CapacityMax int `json:"capacity_max"`
}

type AdminRotateQRTokensRequest struct {
	GateIDs []string `json:"gate_ids"`
}

type AdminStatusUpdateRequest struct {
	Status string `json:"status"`
}
