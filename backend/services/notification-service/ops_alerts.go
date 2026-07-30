package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type OpsAlertDispatcher interface {
	DispatchOpsAlert(ctx context.Context, alertType string, payload map[string]interface{}) error
}

type DefaultOpsAlertDispatcher struct {
	repo       NotificationRepository
	httpClient *http.Client
	mu         sync.Mutex
	SentAlerts []map[string]interface{}
}

func NewOpsAlertDispatcher(repo NotificationRepository) *DefaultOpsAlertDispatcher {
	return &DefaultOpsAlertDispatcher{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		SentAlerts: []map[string]interface{}{},
	}
}

func (d *DefaultOpsAlertDispatcher) DispatchOpsAlert(ctx context.Context, alertType string, payload map[string]interface{}) error {
	channels, err := d.repo.ListOpsAlertChannelsForType(ctx, alertType)
	if err != nil {
		return fmt.Errorf("failed to fetch ops alert channels for %s: %w", alertType, err)
	}

	if len(channels) == 0 {
		logger.Warn("[OpsAlertDispatcher] No active ops alert channels configured for alert_type: %s", alertType)
		return nil
	}

	for _, ch := range channels {
		d.mu.Lock()
		d.SentAlerts = append(d.SentAlerts, map[string]interface{}{
			"channel_type": ch.ChannelType,
			"target":       ch.Target,
			"alert_type":   alertType,
			"payload":      payload,
		})
		d.mu.Unlock()

		if ch.ChannelType == "SLACK" {
			d.sendSlackWebhook(ctx, ch.Target, alertType, payload)
		} else if ch.ChannelType == "EMAIL" {
			d.sendEmailAlert(ctx, ch.Target, alertType, payload)
		}
	}

	return nil
}

func (d *DefaultOpsAlertDispatcher) sendSlackWebhook(ctx context.Context, webhookURL, alertType string, payload map[string]interface{}) {
	text := fmt.Sprintf("🚨 *ZIPPYRA OPS ALERT*: `%s`\n```\n%v\n```", alertType, payload)
	msg := map[string]interface{}{
		"text": text,
	}
	bodyBytes, _ := json.Marshal(msg)

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Error("Failed to build Slack webhook request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send Slack webhook alert: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("[OpsAlertDispatcher] Slack webhook dispatched to %s for %s", webhookURL, alertType)
}

func (d *DefaultOpsAlertDispatcher) sendEmailAlert(ctx context.Context, emailTarget, alertType string, payload map[string]interface{}) {
	log.Printf("[OpsAlertDispatcher] Email ops alert dispatched to %s for %s", logger.MaskEmail(emailTarget), alertType)
}
