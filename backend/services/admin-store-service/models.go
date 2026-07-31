package main

import (
	"time"
)

const (
	ChainStatusActive    = "ACTIVE"
	ChainStatusSuspended = "SUSPENDED"

	CodeChainNotFound  = "CHAIN_NOT_FOUND"
	CodeChainSuspended = "CHAIN_SUSPENDED"

	StoreStatusActive           = "ACTIVE"
	StoreStatusInactive         = "INACTIVE"
	StoreStatusUnderMaintenance = "UNDER_MAINTENANCE"
)

// Chain is the canonical domain model for a retail chain — owned by admin-store-service.
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

// ── Chain Request / Response structs ─────────────────────────────────────────

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

// ── Store Admin Request structs (mirrors store-service models) ─────────────────

type CreateStoreRequest struct {
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

type UpdateGeofenceRequest struct {
	Polygon      []Point `json:"polygon,omitempty"`
	RadiusMeters int     `json:"radius_meters"`
}

type UpdateHoursRequest struct {
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
	Timezone    string `json:"timezone"`
}

type UpdateCapacityRequest struct {
	CapacityMax int `json:"capacity_max"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type UpdatePaymentSetupRequest struct {
	RazorpayAccountID string `json:"razorpay_account_id"`
	RazorpayKYCStatus string `json:"razorpay_kyc_status"`
	PaymentSetupNote  string `json:"payment_setup_note"`
}

type RotateQRTokensRequest struct {
	GateIDs []string `json:"gate_ids"`
}

// Point is a geographic coordinate used in geofence polygons.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// StoreResponse is the store representation forwarded from store-service.
type StoreResponse struct {
	ID                   string    `json:"id"`
	ChainID              string    `json:"chain_id"`
	Name                 string    `json:"name"`
	Address              string    `json:"address"`
	City                 string    `json:"city"`
	State                string    `json:"state"`
	Pincode              string    `json:"pincode"`
	GSTIN                string    `json:"gstin,omitempty"`
	Lat                  float64   `json:"lat"`
	Lng                  float64   `json:"lng"`
	GeofenceRadiusMeters int       `json:"geofence_radius_meters"`
	CapacityMax          int       `json:"capacity_max"`
	OpeningTime          string    `json:"opening_time"`
	ClosingTime          string    `json:"closing_time"`
	Timezone             string    `json:"timezone"`
	Status               string    `json:"status"`
	RFIDEnabled          bool      `json:"rfid_enabled"`
	RazorpayAccountID    string    `json:"razorpay_account_id,omitempty"`
	RazorpayKYCStatus    string    `json:"razorpay_kyc_status,omitempty"`
	PaymentSetupNote     string    `json:"payment_setup_note,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// QRToken mirrors store-service's StoreQRToken for the GetQRTokens passthrough response.
type QRToken struct {
	ID        string    `json:"id"`
	StoreID   string    `json:"store_id"`
	GateID    string    `json:"gate_id"`
	Token     string    `json:"token"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
