package main

import (
	"context"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type OfflineCheckJob struct {
	repo     Repository
	producer *kafka.Producer
}

func NewOfflineCheckJob(repo Repository, producer *kafka.Producer) *OfflineCheckJob {
	return &OfflineCheckJob{
		repo:     repo,
		producer: producer,
	}
}

func (j *OfflineCheckJob) RunCheck(ctx context.Context) int {
	devices, _, err := j.repo.ListDevices(ctx, "", StatusActive, "", 1, 1000)
	if err != nil {
		return 0
	}

	offlineCount := 0
	now := time.Now().UTC()

	for _, d := range devices {
		if d.LastHeartbeatAt == nil || now.Sub(*d.LastHeartbeatAt) > 180*time.Second {
			// Device timed out (>180s)
			_ = j.repo.UpdateDeviceStatus(ctx, d.ID, StatusOffline)
			offlineCount++

			// Check for existing unresolved OFFLINE alert to prevent duplicate alert spam
			existingAlert, _ := j.repo.GetUnresolvedAlert(ctx, d.ID, AlertTypeOffline)
			if existingAlert == nil {
				alert := &DeviceAlert{
					DeviceID:  d.ID,
					StoreID:   d.StoreID,
					AlertType: AlertTypeOffline,
					Detail: map[string]interface{}{
						"last_heartbeat_at": d.LastHeartbeatAt,
						"gate_id":           d.GateID,
						"label":             d.Label,
					},
				}
				_ = j.repo.CreateAlert(ctx, alert)

				if j.producer != nil {
					_ = j.producer.PublishEvent(ctx, "device.offline", d.ID, map[string]interface{}{
						"device_id":  d.ID,
						"store_id":   d.StoreID,
						"gate_id":    d.GateID,
						"timestamp":  now,
						"event_type": "device.offline",
					})
				}
				logger.Warn("[Device Mgmt Alert] Device %s (%s) marked OFFLINE after 180s timeout", d.ID, d.Label)
			}
		}
	}

	return offlineCount
}
