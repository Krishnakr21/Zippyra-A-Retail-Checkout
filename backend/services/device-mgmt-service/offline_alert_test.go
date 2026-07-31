package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestOfflineAlert_TimeoutExceeding180s_MarksOfflineNoDuplicates(t *testing.T) {
	repo := NewMemoryRepository()
	job := NewOfflineCheckJob(repo, nil)
	sub := NewMQTTSubscriber(repo, nil)

	staleTime := time.Now().Add(-200 * time.Second)
	device := &Device{
		ID:              "dev-stale-1",
		StoreID:         "store-100",
		ChainID:         "chain-100",
		DeviceType:      DeviceTypeScanner,
		Label:           "Stale Scanner",
		Status:          StatusActive,
		LastHeartbeatAt: &staleTime,
	}
	_ = repo.CreateDevice(context.Background(), device)

	// First job tick: should mark OFFLINE and create 1 alert
	count1 := job.RunCheck(context.Background())
	if count1 != 1 {
		t.Fatalf("expected 1 device marked offline, got %d", count1)
	}

	alerts, _ := repo.ListAlerts(context.Background(), "store-100", nil)
	if len(alerts) != 1 || alerts[0].AlertType != AlertTypeOffline {
		t.Fatalf("expected 1 OFFLINE alert created, got %d", len(alerts))
	}

	// Second job tick: should not create duplicate alert
	count2 := job.RunCheck(context.Background())
	if count2 != 0 {
		t.Fatalf("expected 0 additional devices marked offline on second tick, got %d", count2)
	}

	alertsAfterTick2, _ := repo.ListAlerts(context.Background(), "store-100", nil)
	if len(alertsAfterTick2) != 1 {
		t.Fatalf("expected no duplicate alert created on second tick, got %d alerts", len(alertsAfterTick2))
	}

	// Heartbeat arrives: device transitions back to ACTIVE and alert is resolved
	payload, _ := json.Marshal(map[string]interface{}{"firmware_version": "v1.0.0"})
	topic := "zippyra/store/store-100/SCANNER/dev-stale-1/heartbeat"
	_ = sub.ProcessHeartbeatMessage(context.Background(), topic, payload)

	updated, _ := repo.GetDeviceByID(context.Background(), "dev-stale-1")
	if updated.Status != StatusActive {
		t.Fatalf("expected device status ACTIVE after heartbeat, got %s", updated.Status)
	}

	resolvedAlert, _ := repo.GetUnresolvedAlert(context.Background(), "dev-stale-1", AlertTypeOffline)
	if resolvedAlert != nil {
		t.Fatalf("expected OFFLINE alert to be resolved after heartbeat")
	}
}
