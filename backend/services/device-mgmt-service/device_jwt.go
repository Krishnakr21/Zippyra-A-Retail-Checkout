package main

import (
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func IssueDeviceJWT(device *Device, secret string) (string, string, error) {
	gateID := ""
	if device.GateID != nil {
		gateID = *device.GateID
	}

	claims := &jwt.Claims{
		UserID:     device.ID,
		StoreID:    device.StoreID,
		ChainID:    device.ChainID,
		UserType:   "DEVICE",
		DeviceType: device.DeviceType,
		GateID:     gateID,
	}

	// 1 year TTL for physical hardware DEVICE JWT
	token, err := jwt.GenerateToken(claims, secret, 365*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	kid := "v1-dev-key"
	return token, kid, nil
}
