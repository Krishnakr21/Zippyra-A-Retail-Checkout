package main

import (
	"time"
)

const (
	RoleCashier        = "CASHIER"
	RoleStockAssociate = "STOCK_ASSOCIATE"
	RoleSecurity       = "SECURITY"
	RoleManager        = "MANAGER"
	RoleAdmin          = "ADMIN"

	AuthMethodOTP = "OTP"
	AuthMethodPIN = "PIN"

	// Custom error codes
	CodeStaffNotRegistered  = "STAFF_NOT_REGISTERED"
	CodePhoneAlreadyStaff   = "PHONE_ALREADY_STAFF"
	CodeStoreScopeMismatch = "STORE_SCOPE_MISMATCH"
	CodeStepUpRequired      = "STEP_UP_REQUIRED"
	CodePinNotSet           = "PIN_NOT_SET"
	CodePinInvalid          = "PIN_INVALID"
	CodePinLocked           = "PIN_LOCKED"
	CodeShiftAlreadyActive  = "SHIFT_ALREADY_ACTIVE"
	CodeNoActiveShift       = "NO_ACTIVE_SHIFT"
	CodeStaffNotFound       = "STAFF_NOT_FOUND"
	CodeInvalidRole         = "INVALID_ROLE"
)

type StaffMember struct {
	ID                 string     `json:"id"`
	StoreID            string     `json:"store_id"`
	ChainID            string     `json:"chain_id"`
	Phone              string     `json:"phone"`
	Name               string     `json:"name"`
	Role               string     `json:"role"`
	PinHash            *string    `json:"-"`
	PinSetAt           *time.Time `json:"pin_set_at,omitempty"`
	PinFailedAttempts  int        `json:"-"`
	PinLockedUntil     *time.Time `json:"pin_locked_until,omitempty"`
	IsActive           bool       `json:"is_active"`
	CreatedBy          *string    `json:"created_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type StaffSession struct {
	ID         string     `json:"id"`
	StaffID    string     `json:"staff_id"`
	DeviceID   string     `json:"device_id"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	AuthMethod string     `json:"auth_method"`
}

type StaffShift struct {
	ID        string     `json:"id"`
	StaffID   string     `json:"staff_id"`
	StoreID   string     `json:"store_id"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	StaffName string     `json:"staff_name,omitempty"`
}

type CreateStaffRequest struct {
	StoreID string `json:"store_id"`
	Phone   string `json:"phone"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}

type UpdateStaffRequest struct {
	Name  *string `json:"name,omitempty"`
	Role  *string `json:"role,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

type SendOtpRequest struct {
	Phone string `json:"phone"`
}

type VerifyOtpRequest struct {
	Phone    string `json:"phone"`
	OTP      string `json:"otp"`
	DeviceID string `json:"device_id"`
}

type SetPinRequest struct {
	PIN string `json:"pin"`
}

type PinLoginRequest struct {
	Phone    string `json:"phone"`
	PIN      string `json:"pin"`
	DeviceID string `json:"device_id"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	Staff        StaffProfile `json:"staff"`
}

type StaffProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	StoreID   string `json:"store_id"`
	StoreName string `json:"store_name"`
	HasPinSet bool   `json:"has_pin_set"`
}
