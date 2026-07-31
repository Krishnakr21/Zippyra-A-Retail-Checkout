-- Migration: 000003_store_service_init.up.sql
-- Store Service schema for stores, store_qr_tokens, and store_sessions

CREATE TABLE IF NOT EXISTS stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    pincode VARCHAR(20) NOT NULL,
    gstin VARCHAR(50),
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    geofence_polygon JSONB NULL,
    geofence_radius_meters INT DEFAULT 100,
    capacity_max INT NOT NULL DEFAULT 50,
    opening_time TIME,
    closing_time TIME,
    timezone VARCHAR(50) DEFAULT 'Asia/Kolkata',
    status VARCHAR(20) DEFAULT 'ACTIVE',
    rfid_enabled BOOLEAN DEFAULT false,
    catalog_version BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS store_qr_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    gate_id VARCHAR(50) NOT NULL,
    token VARCHAR(64) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS store_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    device_id VARCHAR(100),
    bound_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unbound_at TIMESTAMPTZ NULL,
    catalog_version_at_bind BIGINT
);

-- Unique index ensuring only one active (unbound_at IS NULL) session per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_store_sessions_user_active ON store_sessions(user_id) WHERE unbound_at IS NULL;

-- Query performance indexes
CREATE INDEX IF NOT EXISTS idx_stores_chain ON stores(chain_id);
CREATE INDEX IF NOT EXISTS idx_store_sessions_active ON store_sessions(store_id) WHERE unbound_at IS NULL;
