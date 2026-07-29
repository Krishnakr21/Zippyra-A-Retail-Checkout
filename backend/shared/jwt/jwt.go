package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const STEP_UP_FRESHNESS_MINUTES = 10

type Claims struct {
	UserID     string `json:"user_id,omitempty"`
	AdminID    string `json:"admin_id,omitempty"`
	Email      string `json:"email,omitempty"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
	GateID     string `json:"gate_id,omitempty"`
	StoreID    string `json:"store_id,omitempty"`
	ChainID    string `json:"chain_id,omitempty"`
	Role       string `json:"role,omitempty"`
	UserType   string `json:"user_type,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	StepUpAt   int64  `json:"step_up_at,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT access token.
func GenerateAccessToken(userID, deviceID, secret string, ttl time.Duration) (string, error) {
	return GenerateAccessTokenWithSession(userID, deviceID, "", secret, ttl)
}

// GenerateAccessTokenWithSession creates a signed JWT access token containing SessionID.
func GenerateAccessTokenWithSession(userID, deviceID, sessionID, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		DeviceID:  deviceID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "zippyra-auth-service",
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"zippyra-mobile"},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "v1"
	return token.SignedString([]byte(secret))
}

// GenerateToken creates a signed token from a Claims struct.
func GenerateToken(claims *Claims, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    "zippyra-retailer-auth-service",
		Subject:   claims.UserID,
		Audience:  jwt.ClaimStrings{"zippyra-staff"},
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "v1"
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generates a secure random refresh token and returns (rawToken, hashHex, error)
func GenerateRefreshToken() (string, string, error) {
	raw := uuid.New().String() + "-" + uuid.New().String()
	hash := HashRefreshToken(raw)
	return raw, hash, nil
}

// HashRefreshToken hashes a refresh token using SHA-256
func HashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// ParseAndVerifyToken parses and validates a JWT token string
func ParseAndVerifyToken(tokenStr, secret string) (*Claims, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

type SessionClaims struct {
	UserID    string `json:"user_id"`
	StoreID   string `json:"store_id"`
	SessionID string `json:"session_id"`
	UserType  string `json:"user_type"`
	Role      string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// GenerateSessionToken creates a signed store session JWT (4h TTL)
func GenerateSessionToken(userID, storeID, sessionID, userType, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	now := time.Now()
	role := "CUSTOMER"
	if userType == "STAFF" {
		role = "SECURITY"
	}
	claims := SessionClaims{
		UserID:    userID,
		StoreID:   storeID,
		SessionID: sessionID,
		UserType:  userType,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "zippyra-store-service",
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"zippyra-mobile-session"},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "v1"
	return token.SignedString([]byte(secret))
}

// ParseAndVerifySessionToken validates a store session JWT
func ParseAndVerifySessionToken(tokenStr, secret string) (*SessionClaims, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid session token claims")
	}
	return claims, nil
}

type ExitClaims struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	StoreID   string `json:"store_id"`
	SessionID string `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateExitToken(orderID, userID, storeID, sessionID, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	now := time.Now()
	claims := ExitClaims{
		OrderID:   orderID,
		UserID:    userID,
		StoreID:   storeID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "zippyra-order-service",
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"zippyra-exit-gate"},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "v1"
	return token.SignedString([]byte(secret))
}

func ParseAndVerifyExitToken(tokenStr, secret string) (*ExitClaims, bool, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	claims := &ExitClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && claims.OrderID != "" {
			return claims, true, nil
		}
		return nil, false, err
	}

	if !token.Valid {
		return nil, false, errors.New("invalid exit token claims")
	}

	return claims, false, nil
}

type DeviceClaims struct {
	DeviceID string `json:"device_id"`
	GateID   string `json:"gate_id"`
	StoreID  string `json:"store_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

func GenerateDeviceToken(deviceID, gateID, storeID, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	now := time.Now()
	claims := DeviceClaims{
		DeviceID: deviceID,
		GateID:   gateID,
		StoreID:  storeID,
		UserType: "DEVICE",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "zippyra-device-mgmt-service",
			Subject:   deviceID,
			Audience:  jwt.ClaimStrings{"zippyra-gate-device"},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "v1"
	return token.SignedString([]byte(secret))
}

func ParseAndVerifyDeviceToken(tokenStr, secret string) (*DeviceClaims, error) {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	token, err := jwt.ParseWithClaims(tokenStr, &DeviceClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*DeviceClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid device token claims")
	}
	return claims, nil
}


