package main

import (
	"context"
	"testing"
	"time"
)

func TestCertExpiry_ExpiringIn20Days_CreatesAlert(t *testing.T) {
	repo := NewMemoryRepository()
	iot := NewMockIoTProvider()
	job := NewCertExpiryJob(repo, iot)

	expiringSoon := time.Now().AddDate(0, 0, 20)
	expiringFar := time.Now().AddDate(0, 0, 60)

	_ = repo.CreateDevice(context.Background(), &Device{
		ID:            "dev-cert-soon",
		StoreID:       "store-100",
		DeviceType:    DeviceTypeKiosk,
		Label:         "Kiosk 1",
		Status:        StatusActive,
		CertExpiresAt: &expiringSoon,
	})

	_ = repo.CreateDevice(context.Background(), &Device{
		ID:            "dev-cert-far",
		StoreID:       "store-100",
		DeviceType:    DeviceTypeKiosk,
		Label:         "Kiosk 2",
		Status:        StatusActive,
		CertExpiresAt: &expiringFar,
	})

	count := job.RunCheck(context.Background())
	if count != 1 {
		t.Fatalf("expected 1 cert expiry alert created, got %d", count)
	}

	alerts, _ := repo.ListAlerts(context.Background(), "store-100", nil)
	if len(alerts) != 1 || alerts[0].DeviceID != "dev-cert-soon" {
		t.Fatalf("expected alert specifically for dev-cert-soon, got %v", alerts)
	}
}
