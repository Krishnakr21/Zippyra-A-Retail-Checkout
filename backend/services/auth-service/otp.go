package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/redis"
)

type OTPManager interface {
	SendOTP(ctx context.Context, channel, identifier, ip string) (string, error)
	VerifyOTP(ctx context.Context, channel, identifier, code string) error
}

type DefaultOTPManager struct {
	redisClient *redis.Client
	smsSender   SmsSender
	emailSender EmailSender

	// In-memory fallback if Redis is unavailable or for tests
	mu            sync.RWMutex
	memStore      map[string]string        // key -> sha256 hash
	memAttempts   map[string]int           // key -> attempt count
	memLocks      map[string]time.Time     // key -> locked until
	memRateLimits map[string][]time.Time   // key -> request timestamps
}

func NewDefaultOTPManager(rdb *redis.Client, sms SmsSender, email EmailSender) *DefaultOTPManager {
	return &DefaultOTPManager{
		redisClient:   rdb,
		smsSender:     sms,
		emailSender:   email,
		memStore:      make(map[string]string),
		memAttempts:   make(map[string]int),
		memLocks:      make(map[string]time.Time),
		memRateLimits: make(map[string][]time.Time),
	}
}

// Generate6DigitCode generates a 6-digit numeric OTP using crypto/rand
func Generate6DigitCode() (string, error) {
	maxVal := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashOTP hashes OTP code using SHA-256
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

if current > limit then
    return 0
else
    return 1
end
`

func (m *DefaultOTPManager) SendOTP(ctx context.Context, channel, identifier, ip string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// 1. Check Rate Limit (5 per 15 min per identifier, 10 per 15 min per IP)
	if err := m.checkRateLimit(ctx, channel, identifier, ip); err != nil {
		return "", err
	}

	// 2. Check if locked out due to wrong attempts
	if m.isLockedOut(ctx, channel, identifier) {
		return "", errors.NewAPIError(errors.CodeOTPLocked, "Too many failed attempts. Account locked for 15 minutes", nil)
	}

	// 3. Generate 6-digit OTP code using crypto/rand
	code, err := Generate6DigitCode()
	if err != nil {
		return "", errors.NewAPIError(errors.CodeInternalError, "Failed to generate OTP code", nil)
	}
	codeHash := HashOTP(code)

	// 4. Store code hash in Redis / Memory (TTL 300s = 5 minutes)
	otpKey := fmt.Sprintf("otp:%s:%s", channel, identifier)
	attemptsKey := fmt.Sprintf("otp:%s:%s:attempts", channel, identifier)

	if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
		m.redisClient.Set(ctx, otpKey, codeHash, 300*time.Second)
		m.redisClient.Set(ctx, attemptsKey, 0, 300*time.Second)
	} else {
		m.mu.Lock()
		m.memStore[otpKey] = codeHash
		m.memAttempts[attemptsKey] = 0
		m.mu.Unlock()
	}

	// 5. Deliver via channel
	if channel == "phone" {
		if err := m.smsSender.SendSMS(ctx, identifier, code); err != nil {
			logger.Error("Failed to deliver SMS to %s: %v", logger.MaskPhone(identifier), err)
		}
	} else {
		if err := m.emailSender.SendEmailOTP(ctx, identifier, code); err != nil {
			logger.Error("Failed to deliver Email to %s: %v", logger.MaskEmail(identifier), err)
		}
	}

	return code, nil
}

func (m *DefaultOTPManager) VerifyOTP(ctx context.Context, channel, identifier, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	otpKey := fmt.Sprintf("otp:%s:%s", channel, identifier)
	attemptsKey := fmt.Sprintf("otp:%s:%s:attempts", channel, identifier)
	lockKey := fmt.Sprintf("otp:lock:%s:%s", channel, identifier)

	// 1. Check if locked out
	if m.isLockedOut(ctx, channel, identifier) {
		return errors.NewAPIError(errors.CodeOTPLocked, "Too many failed attempts. Account locked for 15 minutes", nil)
	}

	// 2. Retrieve stored Hash
	var storedHash string
	var attempts int

	if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
		val, err := m.redisClient.Get(ctx, otpKey).Result()
		if err != nil || val == "" {
			return errors.NewAPIError(errors.CodeOTPExpired, "OTP has expired or does not exist", nil)
		}
		storedHash = val
		att, _ := m.redisClient.Get(ctx, attemptsKey).Int()
		attempts = att
	} else {
		m.mu.RLock()
		val, ok := m.memStore[otpKey]
		attempts = m.memAttempts[attemptsKey]
		m.mu.RUnlock()
		if !ok || val == "" {
			return errors.NewAPIError(errors.CodeOTPExpired, "OTP has expired or does not exist", nil)
		}
		storedHash = val
	}

	// 3. Compare hash using constant-time comparison
	inputHash := HashOTP(code)
	match := subtle.ConstantTimeCompare([]byte(storedHash), []byte(inputHash)) == 1

	if !match {
		attempts++
		if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
			m.redisClient.Set(ctx, attemptsKey, attempts, 300*time.Second)
		} else {
			m.mu.Lock()
			m.memAttempts[attemptsKey] = attempts
			m.mu.Unlock()
		}

		if attempts >= 5 {
			// Lock for 15 minutes (900 seconds)
			if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
				m.redisClient.Set(ctx, lockKey, "locked", 900*time.Second)
				m.redisClient.Del(ctx, otpKey, attemptsKey)
			} else {
				m.mu.Lock()
				m.memLocks[lockKey] = time.Now().Add(900 * time.Second)
				delete(m.memStore, otpKey)
				delete(m.memAttempts, attemptsKey)
				m.mu.Unlock()
			}
			return errors.NewAPIError(errors.CodeOTPLocked, "Too many failed attempts. Account locked for 15 minutes", nil)
		}
		return errors.NewAPIError(errors.CodeOTPInvalid, "Invalid OTP code", nil)
	}

	// OTP matched! Clean up keys
	if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
		m.redisClient.Del(ctx, otpKey, attemptsKey)
	} else {
		m.mu.Lock()
		delete(m.memStore, otpKey)
		delete(m.memAttempts, attemptsKey)
		m.mu.Unlock()
	}

	return nil
}

func (m *DefaultOTPManager) checkRateLimit(ctx context.Context, channel, identifier, ip string) error {
	idKey := fmt.Sprintf("ratelimit:%s:%s", channel, identifier)
	ipKey := fmt.Sprintf("ratelimit:ip:%s", ip)

	if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
		// 5 / 15 min per identifier
		resId, err := m.redisClient.Eval(ctx, rateLimitScript, []string{idKey}, 5, 900).Int()
		if err == nil && resId == 0 {
			return errors.NewAPIError(errors.CodeRateLimitExceeded, "Too many OTP requests for this identifier. Try again in 15 minutes", nil)
		}
		// 10 / 15 min per IP
		resIP, err := m.redisClient.Eval(ctx, rateLimitScript, []string{ipKey}, 10, 900).Int()
		if err == nil && resIP == 0 {
			return errors.NewAPIError(errors.CodeRateLimitExceeded, "Too many OTP requests from this IP address", nil)
		}
	} else {
		m.mu.Lock()
		defer m.mu.Unlock()
		now := time.Now()
		cutoff := now.Add(-15 * time.Minute)

		// Check identifier limit (5)
		idTimes := m.filterTimes(m.memRateLimits[idKey], cutoff)
		if len(idTimes) >= 5 {
			return errors.NewAPIError(errors.CodeRateLimitExceeded, "Too many OTP requests for this identifier. Try again in 15 minutes", nil)
		}
		m.memRateLimits[idKey] = append(idTimes, now)

		// Check IP limit (10)
		if ip != "" {
			ipTimes := m.filterTimes(m.memRateLimits[ipKey], cutoff)
			if len(ipTimes) >= 10 {
				return errors.NewAPIError(errors.CodeRateLimitExceeded, "Too many OTP requests from this IP address", nil)
			}
			m.memRateLimits[ipKey] = append(ipTimes, now)
		}
	}
	return nil
}

func (m *DefaultOTPManager) isLockedOut(ctx context.Context, channel, identifier string) bool {
	lockKey := fmt.Sprintf("otp:lock:%s:%s", channel, identifier)
	if m.redisClient != nil && m.redisClient.Client != nil && m.redisClient.Ping(ctx).Err() == nil {
		val, err := m.redisClient.Get(ctx, lockKey).Result()
		return err == nil && val != ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	until, ok := m.memLocks[lockKey]
	return ok && time.Now().Before(until)
}

func (m *DefaultOTPManager) filterTimes(times []time.Time, cutoff time.Time) []time.Time {
	var res []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			res = append(res, t)
		}
	}
	return res
}
