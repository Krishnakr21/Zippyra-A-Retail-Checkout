package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      = 64 * 1024
	argonIterations  = 1
	argonParallelism = 4
	argonKeyLen      = 32
	argonSaltLen     = 16
)

func HashPin(pin string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(pin), salt, argonIterations, argonMemory, argonParallelism, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonIterations, argonParallelism, b64Salt, b64Hash)

	return encoded, nil
}

func VerifyPin(pin, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	var version int
	var memory, timeCost uint32
	var parallelism uint8

	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parallelism)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	otherHash := argon2.IDKey([]byte(pin), salt, timeCost, memory, parallelism, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, otherHash) == 1
}

type PINService struct {
	repo Repository
}

func NewPINService(repo Repository) *PINService {
	return &PINService{repo: repo}
}

func (s *PINService) SetPin(ctx context.Context, staffID string, pin string, session *StaffSession) error {
	// Must validate session step-up: session auth_method must be OTP & created_at <= 10m
	if session == nil || session.AuthMethod != AuthMethodOTP {
		return fmt.Errorf(CodeStepUpRequired)
	}
	if time.Since(session.CreatedAt) > 10*time.Minute {
		return fmt.Errorf(CodeStepUpRequired)
	}

	if len(pin) < 4 || len(pin) > 6 {
		return fmt.Errorf("invalid pin format: must be 4-6 digits")
	}

	encoded, err := HashPin(pin)
	if err != nil {
		return fmt.Errorf("failed to hash pin: %w", err)
	}

	return s.repo.UpdateStaffPin(ctx, staffID, encoded)
}

func (s *PINService) VerifyPinLogin(ctx context.Context, phone, pin string) (*StaffMember, error) {
	staff, err := s.repo.GetStaffByPhone(ctx, phone)
	if err != nil || staff == nil || !staff.IsActive || staff.PinHash == nil {
		return nil, fmt.Errorf(CodePinNotSet)
	}

	now := time.Now()
	// Check if currently locked
	if staff.PinLockedUntil != nil && staff.PinLockedUntil.After(now) {
		return nil, fmt.Errorf(CodePinLocked)
	}

	// Verify PIN
	if !VerifyPin(pin, *staff.PinHash) {
		attempts := staff.PinFailedAttempts + 1
		var lockedUntil *time.Time
		if attempts >= 5 {
			lockTime := now.Add(15 * time.Minute)
			lockedUntil = &lockTime
			attempts = 0 // Reset attempts for next window after lock
			_ = s.repo.UpdatePinLockout(ctx, staff.ID, attempts, lockedUntil)
			return nil, fmt.Errorf(CodePinLocked)
		}
		_ = s.repo.UpdatePinLockout(ctx, staff.ID, attempts, nil)
		return nil, fmt.Errorf(CodePinInvalid)
	}

	// Success -> Reset failed attempts
	if staff.PinFailedAttempts > 0 || staff.PinLockedUntil != nil {
		_ = s.repo.UpdatePinLockout(ctx, staff.ID, 0, nil)
	}

	return staff, nil
}
