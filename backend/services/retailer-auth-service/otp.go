package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/zippyra/backend/shared/redis"
	"github.com/zippyra/backend/shared/sms"
)

type OTPService struct {
	repo      Repository
	redis     *redis.Client
	smsSender sms.SmsSender
}

func NewOTPService(repo Repository, redisClient *redis.Client, smsSender sms.SmsSender) *OTPService {
	if smsSender == nil {
		smsSender = &sms.LogSmsSender{}
	}
	return &OTPService{
		repo:      repo,
		redis:     redisClient,
		smsSender: smsSender,
	}
}

func (s *OTPService) checkRateLimit(ctx context.Context, key string, maxLimit int64, ttl time.Duration) bool {
	if s.redis == nil {
		return true
	}
	count := s.redis.Incr(ctx, key).Val()
	if count == 1 {
		s.redis.Expire(ctx, key, ttl)
	}
	return count <= maxLimit
}

func (s *OTPService) SendOTP(ctx context.Context, phone, ip string) error {
	// Rate Limiting (5 per 15 min per phone, 10 per 15 min per IP)
	if s.redis != nil {
		phoneLimitKey := fmt.Sprintf("staff_otp_send_ratelimit:%s", phone)
		ipLimitKey := fmt.Sprintf("staff_otp_send_ratelimit:ip:%s", ip)

		if !s.checkRateLimit(ctx, phoneLimitKey, 5, 15*time.Minute) {
			return fmt.Errorf("RATE_LIMIT_EXCEEDED")
		}
		if !s.checkRateLimit(ctx, ipLimitKey, 10, 15*time.Minute) {
			return fmt.Errorf("RATE_LIMIT_EXCEEDED")
		}
	}

	// 2. Staff check: phone MUST exist and be active in staff_members
	staff, err := s.repo.GetStaffByPhone(ctx, phone)
	if err != nil || staff == nil || !staff.IsActive {
		// Deliberately vague error externally - zero calls to SmsSender
		return fmt.Errorf(CodeStaffNotRegistered)
	}

	// Generate 6-digit OTP via crypto/rand
	otpInt, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Errorf("failed to generate crypto rand OTP: %w", err)
	}
	otpCode := fmt.Sprintf("%06d", otpInt.Int64())

	// Hash code via SHA-256 for Redis storage
	hashBytes := sha256.Sum256([]byte(otpCode))
	hashedCode := hex.EncodeToString(hashBytes[:])

	if s.redis != nil {
		otpKey := fmt.Sprintf("otp:staff:%s", phone)
		attemptsKey := fmt.Sprintf("otp:staff:%s:attempts", phone)

		if err := s.redis.Set(ctx, otpKey, hashedCode, 5*time.Minute).Err(); err != nil {
			return fmt.Errorf("redis error storing OTP: %w", err)
		}
		_ = s.redis.Set(ctx, attemptsKey, "0", 5*time.Minute).Err()
	}

	// Deliver SMS
	return s.smsSender.SendSMS(ctx, phone, otpCode)
}

func (s *OTPService) VerifyOTP(ctx context.Context, phone, otpCode string) (*StaffMember, error) {
	staff, err := s.repo.GetStaffByPhone(ctx, phone)
	if err != nil || staff == nil || !staff.IsActive {
		return nil, fmt.Errorf(CodeStaffNotRegistered)
	}

	if s.redis != nil {
		otpKey := fmt.Sprintf("otp:staff:%s", phone)
		attemptsKey := fmt.Sprintf("otp:staff:%s:attempts", phone)

		attempts := s.redis.Incr(ctx, attemptsKey).Val()
		if attempts > 5 {
			_ = s.redis.Del(ctx, otpKey)
			return nil, fmt.Errorf("TOO_MANY_ATTEMPTS")
		}

		storedHash, err := s.redis.Get(ctx, otpKey).Result()
		if err != nil || storedHash == "" {
			return nil, fmt.Errorf("OTP_EXPIRED")
		}

		inputHashBytes := sha256.Sum256([]byte(otpCode))
		inputHash := hex.EncodeToString(inputHashBytes[:])

		if storedHash != inputHash {
			return nil, fmt.Errorf("INVALID_OTP")
		}

		// OTP correct -> clean Redis keys
		_ = s.redis.Del(ctx, otpKey)
		_ = s.redis.Del(ctx, attemptsKey)
	}

	return staff, nil
}
