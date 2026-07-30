package main

import (
	"context"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type CertExpiryJob struct {
	repo Repository
	iot  IoTProvider
}

func NewCertExpiryJob(repo Repository, iot IoTProvider) *CertExpiryJob {
	return &CertExpiryJob{
		repo: repo,
		iot:  iot,
	}
}

func (j *CertExpiryJob) RunCheck(ctx context.Context) int {
	devices, _, err := j.repo.ListDevices(ctx, "", "", "", 1, 1000)
	if err != nil {
		return 0
	}

	alertCount := 0
	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, 30) // Certs expiring in 30 days

	for _, d := range devices {
		if d.Status == StatusDecommissioned {
			continue
		}

		if d.CertExpiresAt != nil && d.CertExpiresAt.Before(cutoff) {
			existingAlert, _ := j.repo.GetUnresolvedAlert(ctx, d.ID, AlertTypeCertExpiringSoon)
			if existingAlert == nil {
				alert := &DeviceAlert{
					DeviceID:  d.ID,
					StoreID:   d.StoreID,
					AlertType: AlertTypeCertExpiringSoon,
					Detail: map[string]interface{}{
						"cert_id":         d.CertID,
						"cert_expires_at": d.CertExpiresAt,
					},
				}
				_ = j.repo.CreateAlert(ctx, alert)
				alertCount++
				logger.Warn("[Device Mgmt Alert] Certificate for device %s (%s) expires soon on %v", d.ID, d.Label, d.CertExpiresAt)
			}
		}
	}

	return alertCount
}
