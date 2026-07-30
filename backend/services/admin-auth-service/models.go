package main

import (
	"time"
)

const (
	RoleSuperAdmin    = "SUPER_ADMIN"
	RolePlatformAdmin = "PLATFORM_ADMIN"
	RoleSupport       = "SUPPORT"

	CodeDomainNotAllowed     = "DOMAIN_NOT_ALLOWED"
	CodeAdminNotRegistered   = "ADMIN_NOT_REGISTERED"
	CodeTOTPInvalid          = "TOTP_INVALID"
	CodeTOTPLocked           = "TOTP_LOCKED"
	CodeStepUpRequired       = "STEP_UP_REQUIRED"
	CodeAdminAlreadyExists   = "ADMIN_ALREADY_EXISTS"
	CodeCannotDeactivateSelf = "CANNOT_DEACTIVATE_SELF"
)

type AdminUser struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	Role                string     `json:"role"`
	GoogleSub           *string    `json:"google_sub,omitempty"`
	TOTPSecretEncrypted []byte     `json:"-"`
	TOTPEnabledAt       *time.Time `json:"totp_enabled_at,omitempty"`
	TOTPFailedAttempts  int        `json:"-"`
	TOTPLockedUntil     *time.Time `json:"totp_locked_until,omitempty"`
	IsActive            bool       `json:"is_active"`
	CreatedBy           *string    `json:"created_by,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type AdminSession struct {
	ID        string     `json:"id"`
	AdminID   string     `json:"admin_id"`
	DeviceID  string     `json:"device_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type GoogleLoginRequest struct {
	IDToken  string `json:"id_token"`
	DeviceID string `json:"device_id,omitempty"`
}

type TOTPConfirmRequest struct {
	Code string `json:"code"`
}

type TOTPVerifyRequest struct {
	Code string `json:"code"`
}

type StepUpRequest struct {
	Code string `json:"code"`
}

type CreateAdminRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type UpdateAdminRoleRequest struct {
	Role string `json:"role"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
