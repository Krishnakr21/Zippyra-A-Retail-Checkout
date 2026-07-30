package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const (
	ResultOpened       = "OPENED"
	ResultQRAlreadyUsed = "QR_ALREADY_USED"
	ResultQRExpired     = "QR_EXPIRED"
	ResultInvalidToken  = "INVALID_TOKEN"
	ResultWrongStore    = "WRONG_STORE"
	ResultAwaitingRFID  = "AWAITING_RFID"
	ResultNotAwaitingRFID = "NOT_AWAITING_RFID"
	ResultRFIDConfirmed = "RFID_CONFIRMED"
	ResultRFIDTimeout   = "RFID_TIMEOUT"
	ResultStaffOverride = "STAFF_OVERRIDE"
	ResultDenied        = "DENIED"

	TopicExitValidated   = "exit.validated"
	TopicExitDenied      = "exit.denied"
	TopicExitRFIDFailure = "exit.rfid_failure"
)

type ExitAttempt struct {
	ID         string          `json:"id"`
	OrderID    string          `json:"order_id"`
	UserID     string          `json:"user_id"`
	StoreID    string          `json:"store_id"`
	GateID     string          `json:"gate_id"`
	Result     string          `json:"result"`
	IsAlarm    bool            `json:"is_alarm"`
	RFIDTagIDs json.RawMessage `json:"rfid_tag_ids,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type StaffOverride struct {
	ID          string    `json:"id"`
	OrderID     *string   `json:"order_id,omitempty"`
	StoreID     string    `json:"store_id"`
	GateID      string    `json:"gate_id"`
	StaffUserID string    `json:"staff_user_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// Request & Response DTOs
type ValidateExitRequest struct {
	ExitToken string `json:"exit_token"`
}

type ValidateExitResponse struct {
	Result  string `json:"result"`
	OrderID string `json:"order_id,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type RFIDConfirmRequest struct {
	OrderID             string   `json:"order_id"`
	TagIDs              []string `json:"tag_ids"`
	DeactivationSuccess bool     `json:"deactivation_success"`
}

type StaffOverrideRequest struct {
	OrderID string `json:"order_id"`
	GateID  string `json:"gate_id"`
	Reason  string `json:"reason"` // "SCANNER_MALFUNCTION" | "CUSTOMER_ASSISTANCE" | "OTHER"
}

type ExitStatusResponse struct {
	Result string `json:"result"`
	GateID string `json:"gate_id,omitempty"`
}

// Kafka Event Payloads
type ExitValidatedPayload struct {
	OrderID   string    `json:"order_id"`
	SessionID string    `json:"session_id,omitempty"`
	UserID    string    `json:"user_id"`
	StoreID   string    `json:"store_id"`
	GateID    string    `json:"gate_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ExitDeniedPayload struct {
	OrderID   string    `json:"order_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	StoreID   string    `json:"store_id"`
	GateID    string    `json:"gate_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type ExitRFIDFailurePayload struct {
	OrderID   string    `json:"order_id"`
	StoreID   string    `json:"store_id"`
	GateID    string    `json:"gate_id"`
	TagIDs    []string  `json:"tag_ids"`
	Timestamp time.Time `json:"timestamp"`
}

// MQTT Command Payload
type GateMQTTCommand struct {
	Cmd       string    `json:"cmd"` // "OPEN"
	OrderID   string    `json:"order_id"`
	Timestamp time.Time `json:"ts"`
}

// Prometheus Alarm Metrics Counter
type AlarmMetrics struct {
	mu     sync.RWMutex
	counts map[string]int64
}

func NewAlarmMetrics() *AlarmMetrics {
	return &AlarmMetrics{
		counts: make(map[string]int64),
	}
}

func (m *AlarmMetrics) Inc(storeID, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, result)
	m.counts[key]++
}

func (m *AlarmMetrics) Get(storeID, result string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counts[fmt.Sprintf("%s:%s", storeID, result)]
}
