package main

import (
	"encoding/json"
	"time"
)

type ERPType string

const (
	ERPTypeSAP   ERPType = "SAP"
	ERPTypeTally ERPType = "TALLY"
	ERPTypeBusy  ERPType = "BUSY"
)

type IntegrationMode string

const (
	IntegrationModeDirect      IntegrationMode = "DIRECT"
	IntegrationModeAgentPolled IntegrationMode = "AGENT_POLLED"
)

type ConnectionStatus string

const (
	ConnectionStatusPendingSetup ConnectionStatus = "PENDING_SETUP"
	ConnectionStatusActive       ConnectionStatus = "ACTIVE"
	ConnectionStatusPaused       ConnectionStatus = "PAUSED"
	ConnectionStatusError        ConnectionStatus = "ERROR"
)

type ERPConnection struct {
	ID                            string           `json:"id"`
	ChainID                       string           `json:"chain_id"`
	ERPType                       ERPType          `json:"erp_type"`
	IntegrationMode               IntegrationMode  `json:"integration_mode"`
	DisplayName                   string           `json:"display_name"`
	InboundWebhookSecretEncrypted []byte           `json:"-"`
	AgentAPIKeyHash               *string          `json:"-"`
	OutboundConfigEncrypted       []byte           `json:"-"`
	EnabledOutboundEvents         []string         `json:"enabled_outbound_events"`
	Status                        ConnectionStatus `json:"status"`
	LastInboundAt                 *time.Time       `json:"last_inbound_at,omitempty"`
	LastOutboundAt                *time.Time       `json:"last_outbound_at,omitempty"`
	LastAgentPollAt               *time.Time       `json:"last_agent_poll_at,omitempty"`
	CreatedBy                     string           `json:"created_by"`
	CreatedAt                     time.Time        `json:"created_at"`
	UpdatedAt                     time.Time        `json:"updated_at"`
}

type OutboundConfig struct {
	BaseURL  string `json:"base_url"`
	AuthType string `json:"auth_type"` // BASIC | API_KEY
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

type ProcessingResult string

const (
	ProcessingResultPending  ProcessingResult = "PENDING"
	ProcessingResultApplied  ProcessingResult = "APPLIED"
	ProcessingResultFailed   ProcessingResult = "FAILED"
	ProcessingResultRejected ProcessingResult = "REJECTED"
)

type ERPWebhookEvent struct {
	ID               string           `json:"id"`
	ConnectionID     string           `json:"connection_id"`
	EventID          string           `json:"event_id"`
	EventType        string           `json:"event_type"` // PRICE_UPDATE | STOCK_ADJUSTMENT | GRN_RECEIVED
	RawPayload       json.RawMessage  `json:"raw_payload"`
	ProcessingResult ProcessingResult `json:"processing_result"`
	FailureReason    *string          `json:"failure_reason,omitempty"`
	ProcessedAt      *time.Time       `json:"processed_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

type SyncJobStatus string

const (
	SyncJobStatusPending      SyncJobStatus = "PENDING"
	SyncJobStatusDelivered    SyncJobStatus = "DELIVERED"
	SyncJobStatusAcknowledged SyncJobStatus = "ACKNOWLEDGED"
	SyncJobStatusFailed       SyncJobStatus = "FAILED"
)

type ERPSyncJob struct {
	ID              string        `json:"id"`
	ConnectionID    string        `json:"connection_id"`
	Direction       string        `json:"direction"` // OUTBOUND
	SourceEventType string        `json:"source_event_type"`
	SourceEventID   string        `json:"source_event_id"`
	Payload         json.RawMessage `json:"payload"`
	Status          SyncJobStatus `json:"status"`
	AttemptCount    int           `json:"attempt_count"`
	FailureReason   *string       `json:"failure_reason,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	DeliveredAt     *time.Time    `json:"delivered_at,omitempty"`
	AcknowledgedAt  *time.Time    `json:"acknowledged_at,omitempty"`
}

// Request & Response DTOs

type CreateConnectionRequest struct {
	ChainID               string          `json:"chain_id"`
	ERPType               ERPType         `json:"erp_type"`
	IntegrationMode       IntegrationMode `json:"integration_mode"`
	DisplayName           string          `json:"display_name"`
	OutboundConfig        *OutboundConfig `json:"outbound_config,omitempty"`
	EnabledOutboundEvents []string        `json:"enabled_outbound_events"`
}

type CreateConnectionResponse struct {
	Connection          *ERPConnection `json:"connection"`
	PlaintextSecret     string         `json:"webhook_secret"`
	PlaintextAgentAPIKey *string       `json:"agent_api_key,omitempty"`
	ConnectorSetupNote  string         `json:"connector_setup_note,omitempty"`
}

type UpdateConnectionRequest struct {
	DisplayName           string           `json:"display_name"`
	EnabledOutboundEvents []string         `json:"enabled_outbound_events"`
	Status                ConnectionStatus `json:"status"`
}

type RotateSecretResponse struct {
	PlaintextSecret     string  `json:"webhook_secret"`
	PlaintextAgentAPIKey *string `json:"agent_api_key,omitempty"`
	GracePeriodSeconds  int     `json:"grace_period_seconds"`
}

type AckPullQueueRequest struct {
	JobIDs []string `json:"job_ids"`
}
