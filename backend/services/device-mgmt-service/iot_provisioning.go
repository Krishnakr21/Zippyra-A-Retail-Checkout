package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type IoTCredentialsBundle struct {
	CertARN       string
	CertID        string
	CertPEM       string
	PrivateKeyPEM string
	RootCAPEM     string
	MQTTEndpoint  string
	ExpiresAt     time.Time
}

type IoTProvider interface {
	CreateThingAndCert(ctx context.Context, thingName, deviceType, storeID, deviceID string) (*IoTCredentialsBundle, error)
	DecommissionThingAndCert(ctx context.Context, thingName, certID string) error
	RotateCert(ctx context.Context, thingName, oldCertID, deviceType, storeID, deviceID string) (*IoTCredentialsBundle, error)
	ListCertificatesExpiringSoon(ctx context.Context, days int) ([]string, error)
}

type MockIoTProvider struct{}

func NewMockIoTProvider() *MockIoTProvider {
	return &MockIoTProvider{}
}

func (m *MockIoTProvider) CreateThingAndCert(ctx context.Context, thingName, deviceType, storeID, deviceID string) (*IoTCredentialsBundle, error) {
	if strings.Contains(thingName, "FAIL") || deviceType == "FAIL_TRIGGER" || deviceID == "fail-id" {
		return nil, fmt.Errorf("simulated AWS IoT provisioning failure")
	}

	certID := fmt.Sprintf("cert-id-%s", deviceID[:8])
	certARN := fmt.Sprintf("arn:aws:iot:ap-south-1:123456789012:cert/%s", certID)
	expiresAt := time.Now().AddDate(1, 0, 0) // 1 year cert validity

	return &IoTCredentialsBundle{
		CertARN:       certARN,
		CertID:        certID,
		CertPEM:       "-----BEGIN CERTIFICATE-----\nMOCK_DEVICE_CERTIFICATE_PEM\n-----END CERTIFICATE-----",
		PrivateKeyPEM: "-----BEGIN RSA PRIVATE KEY-----\nMOCK_DEVICE_PRIVATE_KEY_PEM\n-----END RSA PRIVATE KEY-----",
		RootCAPEM:     "-----BEGIN CERTIFICATE-----\nMOCK_AMAZON_ROOT_CA_1\n-----END CERTIFICATE-----",
		MQTTEndpoint:  "a1b2c3d4e5f6g7-ats.iot.ap-south-1.amazonaws.com",
		ExpiresAt:     expiresAt,
	}, nil
}

func (m *MockIoTProvider) DecommissionThingAndCert(ctx context.Context, thingName, certID string) error {
	return nil
}

func (m *MockIoTProvider) RotateCert(ctx context.Context, thingName, oldCertID, deviceType, storeID, deviceID string) (*IoTCredentialsBundle, error) {
	return m.CreateThingAndCert(ctx, thingName, deviceType, storeID, deviceID)
}

func (m *MockIoTProvider) ListCertificatesExpiringSoon(ctx context.Context, days int) ([]string, error) {
	return []string{}, nil
}
