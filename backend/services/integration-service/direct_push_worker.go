package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/crypto"
)

type DirectPushWorker struct {
	repo       IntegrationRepository
	httpClient *http.Client
	masterKey  string
}

func NewDirectPushWorker(repo IntegrationRepository, masterKey string) *DirectPushWorker {
	return &DirectPushWorker{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // 15s timeout for external ERP APIs
		},
		masterKey: masterKey,
	}
}

func (w *DirectPushWorker) PushSyncJob(ctx context.Context, job *ERPSyncJob, conn *ERPConnection) error {
	if conn == nil {
		c, err := w.repo.GetConnectionByID(ctx, job.ConnectionID)
		if err != nil {
			return err
		}
		conn = c
	}

	if conn.IntegrationMode != IntegrationModeDirect {
		return fmt.Errorf("cannot direct push for non-DIRECT connection mode")
	}

	job.AttemptCount++

	var outboundCfg OutboundConfig
	if len(conn.OutboundConfigEncrypted) > 0 {
		dec, err := crypto.Decrypt(conn.OutboundConfigEncrypted, w.masterKey)
		if err == nil {
			_ = json.Unmarshal(dec, &outboundCfg)
		}
	}

	targetURL := outboundCfg.BaseURL
	if targetURL == "" {
		targetURL = "http://localhost:8080/mock-sap-odata"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(job.Payload))
	if err != nil {
		failMsg := fmt.Sprintf("Failed to create HTTP request: %v", err)
		_ = w.repo.UpdateSyncJobStatus(ctx, job.ID, SyncJobStatusFailed, job.AttemptCount, &failMsg)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if outboundCfg.AuthType == "BASIC" && outboundCfg.Username != "" {
		req.SetBasicAuth(outboundCfg.Username, outboundCfg.Password)
	} else if outboundCfg.AuthType == "API_KEY" && outboundCfg.APIKey != "" {
		req.Header.Set("X-API-Key", outboundCfg.APIKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		statusStr := "HTTP_ERROR"
		if resp != nil {
			statusStr = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		failMsg := fmt.Sprintf("SAP Push failed: %s (err: %v)", statusStr, err)
		_ = w.repo.UpdateSyncJobStatus(ctx, job.ID, SyncJobStatusFailed, job.AttemptCount, &failMsg)
		return fmt.Errorf("%s", failMsg)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	// Success -> DELIVERED
	_ = w.repo.UpdateSyncJobStatus(ctx, job.ID, SyncJobStatusDelivered, job.AttemptCount, nil)
	now := time.Now()
	_ = w.repo.UpdateConnectionTimestamps(ctx, conn.ID, nil, &now, nil, nil)
	return nil
}

func (w *DirectPushWorker) StartBackgroundRetryLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runRetrySweep(ctx)
		}
	}
}

func (w *DirectPushWorker) runRetrySweep(ctx context.Context) {
	jobs, err := w.repo.ListFailedDirectSyncJobs(ctx, 10) // retry ceiling = 10
	if err != nil || len(jobs) == 0 {
		return
	}

	log.Printf("[DirectPushWorker] Retrying %d failed DIRECT sync jobs...", len(jobs))
	for _, job := range jobs {
		conn, err := w.repo.GetConnectionByID(ctx, job.ConnectionID)
		if err != nil || conn.Status != ConnectionStatusActive {
			continue
		}
		_ = w.PushSyncJob(ctx, job, conn)
	}
}
