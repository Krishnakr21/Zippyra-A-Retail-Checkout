CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    device_type VARCHAR(30) NOT NULL, -- GATE | RFID_PAD | SCANNER | KIOSK | PRINTER | CAMERA | NFC_READER | DISPLAY
    gate_id VARCHAR(50) NULL,
    label VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PROVISIONING', -- PROVISIONING | ACTIVE | OFFLINE | DECOMMISSIONED
    iot_thing_name VARCHAR(150) UNIQUE NOT NULL,
    cert_arn TEXT NOT NULL,
    cert_id VARCHAR(100) NOT NULL,
    cert_expires_at TIMESTAMPTZ NULL,
    device_jwt_kid VARCHAR(50) NOT NULL,
    last_heartbeat_at TIMESTAMPTZ NULL,
    firmware_version VARCHAR(30) NULL,
    provisioned_at TIMESTAMPTZ NULL,
    decommissioned_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS device_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    store_id UUID NOT NULL,
    alert_type VARCHAR(30) NOT NULL, -- OFFLINE | CERT_EXPIRING_SOON | LOW_BATTERY | HEARTBEAT_ANOMALY
    detail JSONB NULL,
    resolved_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_store_status ON devices(store_id, status);
CREATE INDEX IF NOT EXISTS idx_device_alerts_unresolved ON device_alerts(store_id) WHERE resolved_at IS NULL;
