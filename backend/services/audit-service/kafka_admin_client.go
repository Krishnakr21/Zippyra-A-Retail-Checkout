package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/zippyra/backend/shared/kafka"
)

type DLQTopicSummary struct {
	TopicName             string `json:"topic_name"`
	MessageCount          int64  `json:"message_count"`
	OldestMessageAgeSec   int64  `json:"oldest_message_age_seconds"`
}

type DLQMessage struct {
	Topic     string                 `json:"topic"`
	Offset    int64                  `json:"offset"`
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type KafkaAdminClient struct {
	mu           sync.Mutex
	producer     *kafka.Producer
	mockDLQStore map[string][]DLQMessage
}

func NewKafkaAdminClient(brokers string) *KafkaAdminClient {
	return &KafkaAdminClient{
		producer:     kafka.NewProducer(brokers),
		mockDLQStore: make(map[string][]DLQMessage),
	}
}

func (c *KafkaAdminClient) ListDLQTopics(ctx context.Context) ([]DLQTopicSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	var summaries []DLQTopicSummary
	for topic, msgs := range c.mockDLQStore {
		if strings.HasSuffix(topic, ".dlq") {
			var oldestAge int64 = 0
			if len(msgs) > 0 {
				oldestAge = int64(time.Since(msgs[0].Timestamp).Seconds())
			}
			summaries = append(summaries, DLQTopicSummary{
				TopicName:           topic,
				MessageCount:        int64(len(msgs)),
				OldestMessageAgeSec: oldestAge,
			})
		}
	}

	return summaries, nil
}

func (c *KafkaAdminClient) PeekDLQMessages(ctx context.Context, topic string, limit int, discardedOffsets map[int64]bool) ([]DLQMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	msgs, ok := c.mockDLQStore[topic]
	if !ok {
		return []DLQMessage{}, nil
	}

	var result []DLQMessage
	for _, m := range msgs {
		if discardedOffsets[m.Offset] {
			continue // Filter soft-discarded offsets
		}
		result = append(result, m)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func (c *KafkaAdminClient) ReplayDLQMessages(ctx context.Context, dlqTopic string, offsets []int64) (int, []int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	originalTopic := strings.TrimSuffix(dlqTopic, ".dlq")

	c.mu.Lock()
	msgs := c.mockDLQStore[dlqTopic]
	c.mu.Unlock()

	msgMap := make(map[int64]DLQMessage)
	for _, m := range msgs {
		msgMap[m.Offset] = m
	}

	replayed := 0
	var failed []int64

	for _, offset := range offsets {
		msg, ok := msgMap[offset]
		if !ok {
			failed = append(failed, offset)
			continue
		}

		err := c.producer.PublishEvent(ctx, originalTopic, msg.Key, msg.Value)
		if err != nil {
			failed = append(failed, offset)
		} else {
			replayed++
		}
	}

	return replayed, failed, nil
}

func (c *KafkaAdminClient) SeedMockDLQMessage(topic string, offset int64, key string, rawValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var parsedValue interface{} = rawValue
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(rawValue), &js); err == nil {
		parsedValue = js
	}

	msg := DLQMessage{
		Topic:     topic,
		Offset:    offset,
		Key:       key,
		Value:     parsedValue,
		Timestamp: time.Now().Add(-5 * time.Minute),
	}

	c.mockDLQStore[topic] = append(c.mockDLQStore[topic], msg)
}
