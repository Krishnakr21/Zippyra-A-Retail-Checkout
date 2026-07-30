package main

import (
	"context"
	"testing"

	"github.com/zippyra/backend/shared/errors"
)

func TestOTP_WrongOTP6x_Lockedout(t *testing.T) {
	channels := []string{"phone", "email"}
	identifiers := map[string]string{
		"phone": "+919876543210",
		"email": "testuser@example.com",
	}

	for _, channel := range channels {
		t.Run("Lockout_"+channel, func(t *testing.T) {
			otpMgr := NewDefaultOTPManager(nil, &LogSmsSender{}, NewGmailEmailSender())
			identifier := identifiers[channel]
			ctx := context.Background()

			// 1. Send OTP
			code, err := otpMgr.SendOTP(ctx, channel, identifier, "127.0.0.1")
			if err != nil {
				t.Fatalf("unexpected error sending OTP: %v", err)
			}
			if code == "" {
				t.Fatalf("expected non-empty OTP code")
			}

			// 2. Try wrong OTP 5 times
			wrongCode := "000000"
			for i := 1; i <= 4; i++ {
				err := otpMgr.VerifyOTP(ctx, channel, identifier, wrongCode)
				if err == nil {
					t.Fatalf("expected error for wrong OTP attempt %d", i)
				}
				apiErr, ok := err.(*errors.APIError)
				if !ok || apiErr.Code != errors.CodeOTPInvalid {
					t.Fatalf("expected OTP_INVALID error code on attempt %d, got: %v", i, err)
				}
			}

			// 5th wrong attempt should trigger lockout
			err = otpMgr.VerifyOTP(ctx, channel, identifier, wrongCode)
			if err == nil {
				t.Fatalf("expected error on 5th wrong attempt")
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok || apiErr.Code != errors.CodeOTPLocked {
				t.Fatalf("expected OTP_LOCKED error on 5th attempt, got: %v", err)
			}

			// 6th attempt (even with correct code) should be rejected due to lockout
			err = otpMgr.VerifyOTP(ctx, channel, identifier, code)
			if err == nil {
				t.Fatalf("expected error on 6th attempt after lockout")
			}
			apiErr, ok = err.(*errors.APIError)
			if !ok || apiErr.Code != errors.CodeOTPLocked {
				t.Fatalf("expected OTP_LOCKED error on 6th attempt, got: %v", err)
			}
		})
	}
}

func TestOTP_SuccessVerification(t *testing.T) {
	otpMgr := NewDefaultOTPManager(nil, &LogSmsSender{}, NewGmailEmailSender())
	ctx := context.Background()
	phone := "+919999988888"

	code, err := otpMgr.SendOTP(ctx, "phone", phone, "192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = otpMgr.VerifyOTP(ctx, "phone", phone, code)
	if err != nil {
		t.Fatalf("expected successful OTP verification, got: %v", err)
	}

	// Verify that second verification fails because OTP was consumed
	err = otpMgr.VerifyOTP(ctx, "phone", phone, code)
	if err == nil {
		t.Fatalf("expected error for already consumed OTP")
	}
}
