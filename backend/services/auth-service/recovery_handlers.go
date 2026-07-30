package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

type RecoveryIPRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

var recoveryRateLimiter = &RecoveryIPRateLimiter{
	attempts: make(map[string][]time.Time),
}

func (r *RecoveryIPRateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	var valid []time.Time
	for _, t := range r.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= 3 {
		r.attempts[ip] = valid
		return false
	}

	r.attempts[ip] = append(valid, now)
	return true
}

func (h *AuthHandler) HandleSetRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Missing authentication token", nil)
		return
	}

	// Step-up lite freshness check (JWT issued within 10 minutes)
	tokenIssuedAtStr := r.Header.Get("X-Token-Issued-At")
	if tokenIssuedAtStr != "" {
		if iatSec, err := time.Parse(time.RFC3339, tokenIssuedAtStr); err == nil {
			if time.Since(iatSec) > 10*time.Minute {
				errors.WriteError(w, http.StatusUnauthorized, "FRESH_LOGIN_REQUIRED", "Fresh login required to update recovery email", nil)
				return
			}
		}
	}

	var req SetRecoveryEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Email address is required", nil)
		return
	}

	email := strings.TrimSpace(req.Email)
	if err := h.repo.SetRecoveryEmail(r.Context(), userID, email); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update recovery email", nil)
		return
	}

	// Generate and send OTP to recovery email
	if h.otpManager != nil {
		clientIP := r.RemoteAddr
		_, _ = h.otpManager.SendOTP(r.Context(), "email", email, clientIP)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification OTP sent to recovery email",
		"email":   logger.MaskEmail(email),
	})
}



func (h *AuthHandler) HandleConfirmRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Missing authentication token", nil)
		return
	}

	var req ConfirmRecoveryEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.OTP) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "OTP code is required", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil || user.RecoveryEmail == nil {
		errors.WriteError(w, http.StatusBadRequest, "NO_RECOVERY_EMAIL", "No recovery email associated with account", nil)
		return
	}

	if h.otpManager != nil {
		err := h.otpManager.VerifyOTP(r.Context(), "email", *user.RecoveryEmail, strings.TrimSpace(req.OTP))
		if err != nil {
			errors.WriteError(w, http.StatusBadRequest, "INVALID_OTP", "Invalid or expired OTP code", nil)
			return
		}
	}

	if err := h.repo.ConfirmRecoveryEmail(r.Context(), userID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to confirm recovery email", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Recovery email verified successfully",
	})
}

func (h *AuthHandler) HandleInitiateAccountRecovery(w http.ResponseWriter, r *http.Request) {
	// IP rate limiting: 3 attempts per IP per hour
	clientIP := r.RemoteAddr
	if forward := r.Header.Get("X-Forwarded-For"); forward != "" {
		clientIP = strings.Split(forward, ",")[0]
	}
	if !recoveryRateLimiter.Allow(clientIP) {
		errors.WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many recovery attempts. Please try again in an hour.", nil)
		return
	}

	var req InitiateAccountRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.OriginalPhone) == "" || strings.TrimSpace(req.NewPhone) == "" {
		// Generic message on bad input to prevent enumeration
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "If this matches an account, you will receive next steps.",
		})
		return
	}

	origPhone := strings.TrimSpace(req.OriginalPhone)
	newPhone := strings.TrimSpace(req.NewPhone)

	user, err := h.repo.GetUserByPhone(r.Context(), origPhone)
	if err != nil || user == nil {
		// Generic message on non-existent phone
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "If this matches an account, you will receive next steps.",
		})
		return
	}

	requestID := uuid.New().String()
	if user.RecoveryEmailVerifiedAt != nil && user.RecoveryEmail != nil {
		// Recovery Email path
		recReq := &AccountRecoveryRequest{
			ID:                 requestID,
			UserID:             user.ID,
			OriginalIdentifier: origPhone,
			NewIdentifier:      newPhone,
			VerificationMethod: "RECOVERY_EMAIL_OTP",
			Status:             "PENDING",
			CreatedAt:          time.Now(),
		}
		_ = h.repo.CreateRecoveryRequest(r.Context(), recReq)

		if h.otpManager != nil {
			clientIP := r.RemoteAddr
			_, _ = h.otpManager.SendOTP(r.Context(), "email", *user.RecoveryEmail, clientIP)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "If this matches an account, you will receive next steps.",
			"request_id": requestID,
			"method":     "RECOVERY_EMAIL_OTP",
		})
		return
	}

	// No verified recovery email -> Escalate to Support Manual
	supportTicketID := uuid.New().String()
	recReq := &AccountRecoveryRequest{
		ID:                 requestID,
		UserID:             user.ID,
		OriginalIdentifier: origPhone,
		NewIdentifier:      newPhone,
		VerificationMethod: "SUPPORT_MANUAL",
		Status:             "PENDING",
		SupportTicketID:    &supportTicketID,
		CreatedAt:          time.Now(),
	}
	_ = h.repo.CreateRecoveryRequest(r.Context(), recReq)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "If this matches an account, you will receive next steps.",
		"request_id":        requestID,
		"method":            "SUPPORT_MANUAL",
		"support_ticket_id": supportTicketID,
	})
}

func (h *AuthHandler) HandleConfirmAccountRecovery(w http.ResponseWriter, r *http.Request) {
	var req ConfirmAccountRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RequestID) == "" || strings.TrimSpace(req.OTP) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Request ID and OTP are required", nil)
		return
	}

	recReq, err := h.repo.GetRecoveryRequestByID(r.Context(), req.RequestID)
	if err != nil || recReq.Status != "PENDING" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid or already processed recovery request", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), recReq.UserID)
	if err != nil || user.RecoveryEmail == nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "User not found or missing recovery email", nil)
		return
	}

	if h.otpManager != nil {
		err := h.otpManager.VerifyOTP(r.Context(), "email", *user.RecoveryEmail, strings.TrimSpace(req.OTP))
		if err != nil {
			errors.WriteError(w, http.StatusBadRequest, "INVALID_OTP", "Invalid or expired OTP code", nil)
			return
		}
	}

	// Update phone and revoke ALL existing sessions atomically
	err = h.repo.UpdateUserPhoneAndRevokeSessions(r.Context(), recReq.UserID, recReq.NewIdentifier)
	if err != nil {
		if err.Error() == "PHONE_ALREADY_IN_USE" {
			errors.WriteError(w, http.StatusConflict, "PHONE_ALREADY_IN_USE", "New phone number is already registered to a different account", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update phone number", nil)
		return
	}

	_ = h.repo.UpdateRecoveryRequestStatus(r.Context(), recReq.ID, "COMPLETED")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Account phone number updated successfully. Please sign in with your new phone number.",
	})
}
