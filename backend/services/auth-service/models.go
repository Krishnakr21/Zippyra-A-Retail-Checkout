package main

import (
	"time"
)

type User struct {
	ID                     string     `json:"id"`
	Name                   *string    `json:"name,omitempty"`
	Phone                  *string    `json:"phone,omitempty"`
	Email                  *string    `json:"email,omitempty"`
	GoogleSub              *string    `json:"google_sub,omitempty"`
	EmailVerifiedAt        *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt        *time.Time `json:"phone_verified_at,omitempty"`
	AuthProviderLast       *string    `json:"auth_provider_last,omitempty"`
	RecoveryEmail          *string    `json:"recovery_email,omitempty"`
	RecoveryEmailVerifiedAt *time.Time `json:"recovery_email_verified_at,omitempty"`
}

type AccountRecoveryRequest struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	OriginalIdentifier string     `json:"original_identifier"`
	NewIdentifier      string     `json:"new_identifier"`
	VerificationMethod string     `json:"verification_method"` // RECOVERY_EMAIL_OTP | SUPPORT_MANUAL
	Status             string     `json:"status"`              // PENDING | VERIFIED | REJECTED | COMPLETED
	SupportTicketID    *string    `json:"support_ticket_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
}

type SetRecoveryEmailRequest struct {
	Email string `json:"email"`
}

type ConfirmRecoveryEmailRequest struct {
	OTP string `json:"otp"`
}

type InitiateAccountRecoveryRequest struct {
	OriginalPhone string `json:"original_phone"`
	NewPhone      string `json:"new_phone"`
}

type ConfirmAccountRecoveryRequest struct {
	RequestID string `json:"request_id"`
	OTP       string `json:"otp"`
}

type AuthSession struct {
	ID               string     `json:"id"`
	DeviceID         string     `json:"device_id"`
	DeviceLabel      *string    `json:"device_label,omitempty"`
	RefreshTokenHash string     `json:"refresh_token_hash"`
	UserID           string     `json:"user_id"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

type SessionDTO struct {
	ID          string     `json:"id"`
	DeviceID    string     `json:"device_id"`
	DeviceLabel string     `json:"device_label"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsCurrent   bool       `json:"is_current"`
}

type SendOTPRequest struct {
	Channel    string `json:"channel"`    // "phone" | "email"
	Identifier string `json:"identifier"` // phone number or email address
}

type VerifyOTPRequest struct {
	Channel     string `json:"channel"`
	Identifier  string `json:"identifier"`
	OTP         string `json:"otp"`
	DeviceID    string `json:"device_id"`
	DeviceLabel string `json:"device_label,omitempty"`
}

type GoogleOAuthRequest struct {
	IDToken     string `json:"id_token"`
	DeviceID    string `json:"device_id"`
	DeviceLabel string `json:"device_label,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
	IsNewUser    bool    `json:"is_new_user"`
}

type UserDTO struct {
	ID                     string     `json:"id"`
	Name                   *string    `json:"name,omitempty"`
	Phone                  *string    `json:"phone,omitempty"`
	Email                  *string    `json:"email,omitempty"`
	GoogleSub              *string    `json:"google_sub,omitempty"`
	EmailVerifiedAt        *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt        *time.Time `json:"phone_verified_at,omitempty"`
	AuthProviderLast       *string    `json:"auth_provider_last,omitempty"`
	RecoveryEmail          *string    `json:"recovery_email,omitempty"`
	RecoveryEmailVerifiedAt *time.Time `json:"recovery_email_verified_at,omitempty"`
}

func ToUserDTO(u *User) UserDTO {
	return UserDTO{
		ID:                     u.ID,
		Name:                   u.Name,
		Phone:                  u.Phone,
		Email:                  u.Email,
		GoogleSub:              u.GoogleSub,
		EmailVerifiedAt:        u.EmailVerifiedAt,
		PhoneVerifiedAt:        u.PhoneVerifiedAt,
		AuthProviderLast:       u.AuthProviderLast,
		RecoveryEmail:          u.RecoveryEmail,
		RecoveryEmailVerifiedAt: u.RecoveryEmailVerifiedAt,
	}
}

type AppVersion struct {
	Platform            string    `json:"platform"`
	MinSupportedVersion string    `json:"min_supported_version"`
	LatestVersion       string    `json:"latest_version"`
	HardUpdateBelow     string    `json:"hard_update_below"`
	SoftUpdateMessage   *string   `json:"soft_update_message,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type VersionCheckResponse struct {
	UpdateRequired      bool    `json:"update_required"`
	LatestVersion       string  `json:"latest_version"`
	MinSupportedVersion string  `json:"min_supported_version"`
	SoftUpdateMessage   *string `json:"soft_update_message,omitempty"`
}

type UpdateAppVersionRequest struct {
	MinSupportedVersion string  `json:"min_supported_version"`
	LatestVersion       string  `json:"latest_version"`
	HardUpdateBelow     string  `json:"hard_update_below"`
	SoftUpdateMessage   *string `json:"soft_update_message,omitempty"`
}

