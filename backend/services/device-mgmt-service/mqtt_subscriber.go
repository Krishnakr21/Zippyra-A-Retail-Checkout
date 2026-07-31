package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type MQTTSubscriber struct {
	repo     Repository
	producer *kafka.Producer
}

func NewMQTTSubscriber(repo Repository, producer *kafka.Producer) *MQTTSubscriber {
	return &MQTTSubscriber{
		repo:     repo,
		producer: producer,
	}
}

func (s *MQTTSubscriber) ProcessHeartbeatMessage(ctx context.Context, topic string, payload []byte) error {
	// Expected topic format: zippyra/store/{store_id}/{device_type}/{device_id}/heartbeat
	parts := strings.Split(topic, "/")
	if len(parts) < 5 {
		return nil
	}

	storeID := parts[2]
	deviceID := parts[4]

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payload, &payloadMap); err != nil {
		payloadMap = map[string]interface{}{"raw": string(payload)}
	}

	firmware := ""
	if fw, ok := payloadMap["firmware_version"].(string); ok {
		firmware = fw
	}

	now := time.Now().UTC()
	hb := &DeviceHeartbeat{
		DeviceID: deviceID,
		StoreID:  storeID,
		TS:       now,
		Payload:  payloadMap,
	}

	// 1. Insert into TimescaleDB hypertable
	_ = s.repo.InsertHeartbeatHypertable(ctx, hb)

	// 2. Lookup device
	device, err := s.repo.GetDeviceByID(ctx, deviceID)
	if err != nil || device.Status == StatusDecommissioned {
		return nil
	}

	// 3. Update last heartbeat & firmware version
	var fwPtr *string
	if firmware != "" {
		fwPtr = &firmware
	}
	_ = s.repo.UpdateDeviceHeartbeat(ctx, deviceID, now, fwPtr)

	// 4. Handle status transition
	if device.Status == StatusProvisioning {
		logger.Info("[Device Mgmt] First heartbeat received for device %s (Gate: %v). Transitioning PROVISIONING -> ACTIVE", deviceID, device.GateID)
		_ = s.repo.UpdateDeviceStatus(ctx, deviceID, StatusActive)
	} else if device.Status == StatusOffline {
		logger.Info("[Device Mgmt] Heartbeat received for OFFLINE device %s. Transitioning OFFLINE -> ACTIVE", deviceID)
		_ = s.repo.UpdateDeviceStatus(ctx, deviceID, StatusActive)
		_ = s.repo.ResolveDeviceAlerts(ctx, deviceID, AlertTypeOffline)

		if s.producer != nil {
			_ = s.producer.PublishEvent(ctx, "device.back_online", deviceID, map[string]interface{}{
				"device_id":  deviceID,
				"store_id":   storeID,
				"gate_id":    device.GateID,
				"timestamp":  now,
				"event_type": "device.back_online",
			})
		}
	}

	return nil
}
