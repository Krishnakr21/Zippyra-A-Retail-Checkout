package main

import (
	"time"
)

const (
	DeviceTypeGate      = "GATE"
	DeviceTypeRFIDPad   = "RFID_PAD"
	DeviceTypeScanner   = "SCANNER"
	DeviceTypeKiosk     = "KIOSK"
	DeviceTypePrinter   = "PRINTER"
	DeviceTypeCamera    = "CAMERA"
	DeviceTypeNFCReader = "NFC_READER"
	DeviceTypeDisplay   = "DISPLAY"

	StatusProvisioning   = "PROVISIONING"
	StatusActive         = "ACTIVE"
	StatusOffline        = "OFFLINE"
	StatusDecommissioned = "DECOMMISSIONED"

	AlertTypeOffline          = "OFFLINE"
	AlertTypeCertExpiringSoon = "CERT_EXPIRING_SOON"
	AlertTypeLowBattery       = "LOW_BATTERY"
	AlertTypeHeartbeatAnomaly = "HEARTBEAT_ANOMALY"

	CodeDeviceNotFound             = "DEVICE_NOT_FOUND"
	CodeGateIDAlreadyExists        = "GATE_ID_ALREADY_EXISTS"
	CodeDeviceAlreadyDecommissioned = "DEVICE_ALREADY_DECOMMISSIONED"
	CodeProvisioningFailed         = "PROVISIONING_FAILED"
	CodeDeviceRevoked              = "DEVICE_REVOKED"
	CodePairingCodeInvalid         = "PAIRING_CODE_INVALID"
)

type Device struct {
	ID               string     `json:"id"`
	StoreID          string     `json:"store_id"`
	ChainID          string     `json:"chain_id"`
	DeviceType       string     `json:"device_type"`
	GateID           *string    `json:"gate_id,omitempty"`
	Label            string     `json:"label"`
	Status           string     `json:"status"`
	IoTThingName     string     `json:"iot_thing_name"`
	CertARN          string     `json:"cert_arn"`
	CertID           string     `json:"cert_id"`
	CertExpiresAt    *time.Time `json:"cert_expires_at,omitempty"`
	DeviceJWTKid     string     `json:"device_jwt_kid"`
	LastHeartbeatAt  *time.Time `json:"last_heartbeat_at,omitempty"`
	FirmwareVersion  *string    `json:"firmware_version,omitempty"`
	ProvisionedAt    *time.Time `json:"provisioned_at,omitempty"`
	DecommissionedAt *time.Time `json:"decommissioned_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	IsStale          bool       `json:"is_stale"`
}

type DeviceAlert struct {
	ID         string                 `json:"id"`
	DeviceID   string                 `json:"device_id"`
	StoreID    string                 `json:"store_id"`
	AlertType  string                 `json:"alert_type"`
	Detail     map[string]interface{} `json:"detail,omitempty"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

type DeviceHeartbeat struct {
	DeviceID string                 `json:"device_id"`
	StoreID  string                 `json:"store_id"`
	TS       time.Time              `json:"ts"`
	Payload  map[string]interface{} `json:"payload"`
}

type ProvisionDeviceRequest struct {
	StoreID    string  `json:"store_id"`
	ChainID    string  `json:"chain_id,omitempty"`
	DeviceType string  `json:"device_type"`
	GateID     *string `json:"gate_id,omitempty"`
	Label      string  `json:"label"`
}

type ProvisionDeviceResponse struct {
	DeviceID      string `json:"device_id"`
	DeviceJWT     string `json:"device_jwt"`
	CertPEM       string `json:"cert_pem"`
	PrivateKeyPEM string `json:"private_key_pem"`
	RootCAPEM     string `json:"root_ca_pem"`
	MQTTEndpoint  string `json:"mqtt_endpoint"`
}
