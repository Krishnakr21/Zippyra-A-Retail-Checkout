package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zippyra/backend/shared/logger"
)

type MQTTClient interface {
	PublishGateCommand(ctx context.Context, storeID, gateID string, cmd *GateMQTTCommand) error
	IsConnected() bool
}

type MockMQTTClient struct {
	mu           sync.Mutex
	published    []GateCommandRecord
	shouldFail   bool
	disconnected bool
}

type GateCommandRecord struct {
	StoreID string
	GateID  string
	Cmd     *GateMQTTCommand
}

func NewMockMQTTClient() *MockMQTTClient {
	return &MockMQTTClient{
		published: make([]GateCommandRecord, 0),
	}
}

func (m *MockMQTTClient) PublishGateCommand(ctx context.Context, storeID, gateID string, cmd *GateMQTTCommand) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.disconnected {
		return fmt.Errorf("mqtt client disconnected")
	}

	if m.shouldFail {
		return fmt.Errorf("mqtt publish timeout")
	}

	record := GateCommandRecord{
		StoreID: storeID,
		GateID:  gateID,
		Cmd:     cmd,
	}
	m.published = append(m.published, record)
	topic := fmt.Sprintf("zippyra/store/%s/gate/%s/command", storeID, gateID)
	payload, _ := json.Marshal(cmd)
	logger.Info("[MQTT PUBLISH] Topic: %s | Payload: %s", topic, string(payload))

	return nil
}

func (m *MockMQTTClient) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.disconnected
}

func (m *MockMQTTClient) GetPublished() []GateCommandRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]GateCommandRecord, len(m.published))
	copy(copied, m.published)
	return copied
}
