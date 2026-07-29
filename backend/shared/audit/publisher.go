package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

const TopicAdminActionPerformed = "admin.action_performed"

type AdminAuditEvent struct {
	ActorID       string                 `json:"actor_id"`
	ActorName     string                 `json:"actor_name"`
	ActionType    string                 `json:"action_type"`
	TargetType    string                 `json:"target_type"`
	TargetID      string                 `json:"target_id"`
	Payload       map[string]interface{} `json:"payload"`
	SourceService string                 `json:"source_service"`
	RequestID     string                 `json:"request_id"`
	Timestamp     time.Time              `json:"ts"`
}

type Publisher struct {
	producer      *kafka.Producer
	sourceService string
}

func NewPublisher(producer *kafka.Producer, sourceService string) *Publisher {
	return &Publisher{
		producer:      producer,
		sourceService: sourceService,
	}
}

func (p *Publisher) Publish(ctx context.Context, event AdminAuditEvent) error {
	if event.SourceService == "" {
		event.SourceService = p.sourceService
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if p.producer == nil {
		logger.Info("[Audit Log Mock] Published %s event (%s): %s", event.ActionType, event.SourceService, string(payloadBytes))
		return nil
	}

	return p.producer.PublishEvent(ctx, TopicAdminActionPerformed, event.TargetID, event)
}
