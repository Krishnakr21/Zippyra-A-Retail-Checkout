-- Migration 000021: Integration Service Initial Schema

CREATE TABLE IF NOT EXISTS erp_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    erp_type VARCHAR(20) NOT NULL,
    integration_mode VARCHAR(10) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    inbound_webhook_secret_encrypted BYTEA NOT NULL,
    agent_api_key_hash TEXT NULL,
    outbound_config_encrypted BYTEA NULL,
    enabled_outbound_events JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING_SETUP',
    last_inbound_at TIMESTAMPTZ NULL,
    last_outbound_at TIMESTAMPTZ NULL,
    last_agent_poll_at TIMESTAMPTZ NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_erp_connections_chain ON erp_connections(chain_id) WHERE status != 'PENDING_SETUP';

CREATE TABLE IF NOT EXISTS erp_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES erp_connections(id) ON DELETE CASCADE,
    event_id VARCHAR(150) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    raw_payload JSONB NOT NULL,
    processing_result VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    failure_reason TEXT NULL,
    processed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_erp_webhook_events_conn_event UNIQUE (connection_id, event_id)
);

CREATE TABLE IF NOT EXISTS erp_sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES erp_connections(id) ON DELETE CASCADE,
    direction VARCHAR(10) NOT NULL DEFAULT 'OUTBOUND',
    source_event_type VARCHAR(50) NOT NULL,
    source_event_id VARCHAR(150) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count INT NOT NULL DEFAULT 0,
    failure_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMPTZ NULL,
    acknowledged_at TIMESTAMPTZ NULL,
    CONSTRAINT uq_erp_sync_jobs_conn_event UNIQUE (connection_id, source_event_type, source_event_id)
);

CREATE INDEX IF NOT EXISTS idx_erp_sync_jobs_pending ON erp_sync_jobs(connection_id) WHERE status = 'PENDING';
