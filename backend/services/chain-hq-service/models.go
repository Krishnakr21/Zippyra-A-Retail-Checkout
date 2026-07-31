package main

import (
	"time"
)

const (
	RoleOwner      = "OWNER"
	RoleFinance    = "FINANCE"
	RoleOperations = "OPERATIONS"

	CodeChainHQUserNotRegistered = "CHAIN_HQ_USER_NOT_REGISTERED"
	CodeCannotDeactivateSelf     = "CANNOT_DEACTIVATE_SELF"
	CodeChainAlreadyHasOwner     = "CHAIN_ALREADY_HAS_OWNER"
)

type ChainHQUser struct {
	ID        string    `json:"id"`
	ChainID   string    `json:"chain_id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChainHQSession struct {
	ID            string     `json:"id"`
	ChainHQUserID string     `json:"chain_hq_user_id"`
	DeviceID      string     `json:"device_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type ChainBulkImportJob struct {
	ID              string            `json:"id"`
	ChainID         string            `json:"chain_id"`
	PerStoreJobIDs  map[string]string `json:"per_store_job_ids"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type SendOTPRequest struct {
	Phone string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone    string `json:"phone"`
	OTP      string `json:"otp"`
	DeviceID string `json:"device_id,omitempty"`
}

type InviteUserRequest struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type AdminProvisionOwnerRequest struct {
	ChainID string `json:"chain_id"`
	Phone   string `json:"phone"`
	Name    string `json:"name"`
}

type BulkImportRequest struct {
	Target    string   `json:"target"` // all_stores | specific_stores
	StoreIDs  []string `json:"store_ids,omitempty"`
}
