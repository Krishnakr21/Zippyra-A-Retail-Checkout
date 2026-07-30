package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/zippyra/backend/shared/kafka"
)

type EventConsumer struct {
	repo             IntegrationRepository
	directPushWorker *DirectPushWorker
	storeChainMap    map[string]string // store_id -> chain_id lookup cache
}

func NewEventConsumer(repo IntegrationRepository, directPushWorker *DirectPushWorker) *EventConsumer {
	return &EventConsumer{
		repo:             repo,
		directPushWorker: directPushWorker,
		storeChainMap:    make(map[string]string),
	}
}

func (c *EventConsumer) SetStoreChainMapping(storeID, chainID string) {
	c.storeChainMap[storeID] = chainID
}

func (c *EventConsumer) ProcessEvent(ctx context.Context, topic string, msgValue []byte) error {
	var genericPayload struct {
		EventID   string `json:"event_id"`
		OrderID   string `json:"order_id"`
		StoreID   string `json:"store_id"`
		ChainID   string `json:"chain_id"`
		EventType string `json:"event_type"`
	}

	if err := json.Unmarshal(msgValue, &genericPayload); err != nil {
		return err
	}

	sourceEventType := topic
	if genericPayload.EventType != "" {
		sourceEventType = genericPayload.EventType
	}

	sourceEventID := genericPayload.EventID
	if sourceEventID == "" {
		sourceEventID = genericPayload.OrderID
	}

	chainID := genericPayload.ChainID
	if chainID == "" && genericPayload.StoreID != "" {
		chainID = c.storeChainMap[genericPayload.StoreID]
	}

	if chainID == "" {
		// Cannot resolve chain_id; skip
		return nil
	}

	// Fetch connections for chain
	conns, err := c.repo.ListConnectionsByChain(ctx, chainID)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		if conn.Status != ConnectionStatusActive {
			continue
		}

		// Check if event type enabled
		enabled := false
		for _, ev := range conn.EnabledOutboundEvents {
			if ev == sourceEventType {
				enabled = true
				break
			}
		}
		if !enabled {
			continue
		}

		// Enqueue sync job
		job := &ERPSyncJob{
			ConnectionID:    conn.ID,
			Direction:       "OUTBOUND",
			SourceEventType: sourceEventType,
			SourceEventID:   sourceEventID,
			Payload:         msgValue,
			Status:          SyncJobStatusPending,
		}

		created, err := c.repo.CreateSyncJob(ctx, job)
		if err != nil {
			log.Printf("[EventConsumer] Error creating sync job: %v", err)
			continue
		}

		if !created {
			// Duplicate event (Kafka redelivery), ignored
			continue
		}

		// If DIRECT mode -> immediate push
		if conn.IntegrationMode == IntegrationModeDirect && c.directPushWorker != nil {
			go func(j *ERPSyncJob, cn *ERPConnection) {
				_ = c.directPushWorker.PushSyncJob(context.Background(), j, cn)
			}(job, conn)
		}
	}

	return nil
}

func (c *EventConsumer) RegisterKafkaHandlers(consumer *kafka.Consumer) {
	consumer.RegisterHandler("order.completed", func(ctx context.Context, value []byte) error {
		return c.ProcessEvent(ctx, "order.completed", value)
	})
	consumer.RegisterHandler("inventory.stock_updated", func(ctx context.Context, value []byte) error {
		return c.ProcessEvent(ctx, "inventory.stock_updated", value)
	})
	consumer.RegisterHandler("warehouse.grn_completed", func(ctx context.Context, value []byte) error {
		return c.ProcessEvent(ctx, "warehouse.grn_completed", value)
	})
}
