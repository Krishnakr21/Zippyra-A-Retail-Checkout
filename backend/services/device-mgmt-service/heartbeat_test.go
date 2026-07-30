package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHeartbeat_FirstHeartbeat_TransitionsProvisioningToActive(t *testing.T) {
	repo := NewMemoryRepository()
	sub := NewMQTTSubscriber(repo, nil)

	gateID := "GATE_01"
	device := &Device{
		ID:           "dev-001",
		StoreID:      "store-100",
		ChainID:      "chain-100",
		DeviceType:   DeviceTypeGate,
		GateID:       &gateID,
		Label:        "Entrance Gate",
		Status:       StatusProvisioning,
		IoTThingName: "thing-001",
	}
	_ = repo.CreateDevice(context.Background(), device)

	payload, _ := json.Marshal(map[string]interface{}{
		"firmware_version": "v1.2.3",
		"battery_pct":      98,
	})

	topic := "zippyra/store/store-100/GATE/dev-001/heartbeat"
	err := sub.ProcessHeartbeatMessage(context.Background(), topic, payload)
	if err != nil {
		t.Fatalf("unexpected error processing heartbeat: %v", err)
	}

	updated, err := repo.GetDeviceByID(context.Background(), "dev-001")
	if err != nil || updated.Status != StatusActive {
		t.Fatalf("expected device status ACTIVE after first heartbeat, got %s", updated.Status)
	}
	if updated.FirmwareVersion == nil || *updated.FirmwareVersion != "v1.2.3" {
		t.Fatalf("expected firmware version v1.2.3, got %v", updated.FirmwareVersion)
	}
}
