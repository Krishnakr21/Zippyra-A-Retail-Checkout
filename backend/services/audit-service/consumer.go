package main

import (
	"context"
	"encoding/json"

	"github.com/zippyra/backend/shared/logger"
)

type AuditConsumer struct {
	repo Repository
}

func NewAuditConsumer(repo Repository) *AuditConsumer {
	return &AuditConsumer{repo: repo}
}

func (c *AuditConsumer) ProcessMessage(ctx context.Context, key []byte, value []byte) error {
	var action AdminAction
	if err := json.Unmarshal(value, &action); err != nil {
		logger.Error("[Audit Consumer] Failed to unmarshal message: %v", err)
		return err
	}

	if err := c.repo.CreateAction(ctx, &action); err != nil {
		logger.Error("[Audit Consumer] Failed to record admin action: %v", err)
		return err
	}

	return nil
}
