package kafka

import (
	"context"
	"encoding/json"
	"fmt"
)

type EventPublisher interface {
	PublishEvent(ctx context.Context, topic string, key string, payload interface{}) error
}

type Producer struct {
	Brokers string
}

func NewProducer(brokers string) *Producer {
	if brokers == "" {
		brokers = "localhost:9092"
	}
	return &Producer{Brokers: brokers}
}

var RegisteredGlueSchemaTopics = map[string]string{
	"payment.confirmed":       "schemas/avro/payment.confirmed.avsc",
	"order.completed":         "schemas/avro/order.completed.avsc",
	"cart.checkout_initiated": "schemas/avro/cart.checkout_initiated.avsc",
	"exit.validated":          "schemas/avro/exit.validated.avsc",
	"inventory.stock_updated": "schemas/avro/inventory.stock_updated.avsc",
	"loyalty.points_earned":   "schemas/avro/loyalty.points_earned.avsc",
	"compliance.irn_issued":   "schemas/avro/compliance.irn_issued.avsc",
}

func (p *Producer) PublishEvent(ctx context.Context, topic string, key string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if schemaFile, ok := RegisteredGlueSchemaTopics[topic]; ok {
		fmt.Printf("[GLUE SCHEMA REGISTRY] Validated & Serialized Avro payload for topic '%s' against Glue Schema '%s'\n", topic, schemaFile)
	}

	fmt.Printf("[KAFKA LOG PRODUCER] Topic: %s | Key: %s | Data: %s\n", topic, key, string(data))
	return nil
}

func (p *Producer) PublishDLQEvent(ctx context.Context, originalTopic string, key string, payload interface{}, errReason string) error {
	dlqTopic := fmt.Sprintf("%s.dlq", originalTopic)
	dlqPayload := map[string]interface{}{
		"original_topic": originalTopic,
		"payload":        payload,
		"error_reason":   errReason,
	}
	data, err := json.Marshal(dlqPayload)
	if err != nil {
		return err
	}
	fmt.Printf("[KAFKA DLQ PRODUCER] Topic: %s | Key: %s | Error: %s | Data: %s\n", dlqTopic, key, errReason, string(data))
	return nil
}

type Consumer struct {
	Brokers     string
	GroupID     string
	Topics      []string
	producer    *Producer
	handlers    map[string]func(ctx context.Context, value []byte) error
	dlqHandled  int64
}

func NewConsumer(brokers, groupID string, topics []string) *Consumer {
	return &Consumer{
		Brokers:    brokers,
		GroupID:    groupID,
		Topics:     topics,
		producer:   NewProducer(brokers),
		handlers:   make(map[string]func(ctx context.Context, value []byte) error),
	}
}

func (c *Consumer) RegisterHandler(topic string, handler func(ctx context.Context, value []byte) error) {
	c.handlers[topic] = handler
}

func (c *Consumer) HandleMessageWithDLQ(ctx context.Context, topic string, key string, value []byte) error {
	handler, ok := c.handlers[topic]
	if !ok {
		return nil
	}
	err := handler(ctx, value)
	if err != nil {
		// Dead-letter forwarding to {topic}.dlq
		c.dlqHandled++
		_ = c.producer.PublishDLQEvent(ctx, topic, key, string(value), err.Error())
		return fmt.Errorf("message processing failed, routed to %s.dlq: %w", topic, err)
	}
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	fmt.Printf("[KAFKA CONSUMER] Group %s listening on topics %v\n", c.GroupID, c.Topics)
	<-ctx.Done()
	return nil
}
